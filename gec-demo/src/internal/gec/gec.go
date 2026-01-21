// src/internal/gec/server.go
package gec

/*
#cgo CFLAGS:  -I${SRCDIR}/../../native/gec_runtime/include
#cgo LDFLAGS: ${SRCDIR}/../../native/gec_runtime/build/libgec.a -lstdc++ -lm -ldl -ljson-c -lsentencepiece -licuuc -licudata -lonnxruntime -lonnxruntime_providers_shared -lonnxruntime_providers_cuda -lcuda
#include "inference.h"
*/
import "C"
import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"gec-demo/src/internal/print"
	"gec-demo/src/internal/speechtagger"
)

// C Variables
const (
	maxBatch = C.MAX_BATCH_SIZE // 500
	nClasses = C.GIBB_CLASSES   // 4
)

var (
	rePrefix  = regexp.MustCompile(`(?i)^translate English to (german|french|romanian)`) // Regex to match "Translate English to (German|French|Romanian)" case-insensitively
	rePreproc = regexp.MustCompile(`\s*\n+\s*`)

	CountLT        = 0
	ChanCapacity   = 250
	DeviceCount    = 1
	NumGpuChannels = 1
	GecoChannels   []chan WorkItem

	// Gibberish Score Threshholds
	CleanThreshold = 70
	MildThreshold  = 92
	NoiseThreshold = 40
	SaladThreshold = 40

	IgnoreCollisions = false
	DoMisspellings   = false
	DoGibb           = false
)

func init() {
	print.SetLevel(0) // DEBUG
	GecoChannels = make([]chan WorkItem, NumGpuChannels)
	print.Debug("Total GEC Channels: %d", NumGpuChannels)

	cwd, err := os.Getwd()
	if err != nil {
		print.Error("Failed to get current working directory: %v", err)
		return
	}
	cfg_path := cwd + "/src/native/gec_runtime/config/config.json"
	print.Info("Cfg Path: \x1b[93m%s\x1b[0m", cfg_path)

	for i := range NumGpuChannels {
		gpuId := i % DeviceCount
		print.Debug("Creating Geco[%d] for gpu-%v", i, gpuId)

		// Buffered channel with a capacity of `ChanCapacity`
		GecoChannels[i] = make(chan WorkItem, ChanCapacity)
		go ClaimGpu(cfg_path, gpuId, GecoChannels[i])
	}

	// Initialize the parts-of-speech tagging model
	err = speechtagger.InitTaggingModel()
	if err != nil {
		fmt.Printf("ERROR: Failed to initialize TaggerModel: %v\n", err)
		return
	}
}

func ClaimGpu(cfg_path string, gpuId int, ch chan WorkItem) {
	var geco unsafe.Pointer
	c_path := C.CString(cfg_path)
	defer cFree(c_path)

	// Allocate a Geco object for the channel
	geco = C.NewGeco(c_path, 1, C.int(gpuId))
	if geco == nil {
		print.Error("Failed initalizing GECO for gpu:%d", gpuId)
		return
	}
	defer C.FreeGeco(geco)

	for item := range ch {
		res := CorrectGrammar(&geco, gpuId, item.Text, item.AllTexts)
		item.Ch <- res
	}
}

// Get random index for a channel to use in GEC Channels
func PickGecChannel() int {
	maxInd := len(GecoChannels)
	for range NumGpuChannels {
		choice := rand.Intn(maxInd)
		if len(GecoChannels[choice]) < cap(GecoChannels[choice]) {
			return choice
		}
	}
	return -1
}

// Run G.E.C. requests and return results
func MarkupGrammar(text string) (gec_result *GecResponse, err error) {
	var misspells []Misspell
	var differences []Markup
	var gram_result *GrammarResult

	gec_result = &GecResponse{}

	text = CleanText(text)

	if DoMisspellings {
		print.Warning("Spell checker is not yet implemented.")
		// Find the spelling errors
		misspells, err = DirtySpellChecker(text)
		if err != nil {
			return nil, err
		}
		misspells = SpellChecker(misspells, text)
		ViewMisspells(misspells)
	}

	// Run the model to get the grammatically corrected version of the text
	gram_result, err = ProcessGrammar(text)
	if gram_result.Err != nil {
		return nil, fmt.Errorf("error running GEC, %v. Input Text: %q", gram_result.Err, text)
	}
	ViewGibbs(gram_result.GibbScores)

	// Define and clean the grammatically corrected text
	corrected_text := gram_result.CorrectText
	if corrected_text == "" {
		corrected_text = text
	}
	// Add proper whitespace to beginning and end of corrected text
	begSpace, endSpace := getSpaceAround(text)
	corrected_text = begSpace + strings.TrimSpace(corrected_text) + endSpace

	// Find the spelling errors and text differences between the original and corrected text
	print.Debug("FIND_DIFF - Original Text: %q\nCorrected Text: %q", text, corrected_text)
	differences, err = FindDifference(text, corrected_text, misspells)
	if err != nil {
		return nil, fmt.Errorf("error in findDiff.go, %w", err)
	}
	print.Debug("Size of Differences: %v", len(differences))

	// Format data to JSON
	print.Debug("Formatting data to JSON!")
	text_markups, err_chars, profanity_words, err := FormatToJson(text, differences, misspells, gram_result.GibbScores)
	if err != nil {
		return nil, fmt.Errorf("error in FormatToJson(), %w", err)
	}

	gec_result.CorrectedText = corrected_text
	gec_result.TextMarkups = text_markups
	gec_result.GibberishScores = gram_result.GibbScores
	gec_result.CharacterCount = len(text)
	gec_result.ErrorCharacterCount = err_chars
	gec_result.ContainsProfanity = len(profanity_words) > 0
	gec_result.ServiceTime = gram_result.ServiceTime
	return gec_result, err
}

func ProcessGrammar(text string) (*GrammarResult, error) {
	var result GrammarResult

	all_texts := PreprocessText(text)
	if len(all_texts) <= 0 {
		return nil, fmt.Errorf("PreprocessText() returns an empty list")
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("Text is empty without whitespace")
	}

	// Send the text to the GEC channel & wait for the result
	work_item := WorkItem{
		Text:     text,
		AllTexts: all_texts,
		Ch:       make(chan GrammarResult, 1), // Channel for receiving the result
	}

	// Send an item to the 1st channel in the GecoChannels
	chan_index := PickGecChannel()
	if chan_index == -1 {
		return nil, fmt.Errorf("No available GPU to run the GEC server")
	}
	print.Debug("Sending work item to Chan[%d]", chan_index)
	GecoChannels[chan_index] <- work_item

	// Wait for the result from the channel
	result = <-work_item.Ch
	close(work_item.Ch)
	if result.Err != nil {
		return nil, result.Err
	}
	return &result, nil
}

func CorrectGrammar(geco *unsafe.Pointer, gpuId int, text string, all_texts []string) GrammarResult {
	chanTime := time.Now()
	gram_result := GrammarResult{
		CorrectText: "",
		GibbScores:  []GibbResults{},
		GpuId:       gpuId,
		Err:         nil,
		ServiceTime: 0.0,
	}

	// Check if a pointer is nil
	if geco == nil {
		gram_result.Err = fmt.Errorf("Geco pointer is nil")
		return gram_result
	}

	// Convert Go strings to C strings
	cTexts, ctext_cleanup := goStringsToC(all_texts)
	defer ctext_cleanup()

	// Run grammar correction
	var c_output *C.char
	C.GecoRun(*geco, &cTexts[0], C.int(len(all_texts)), &c_output)
	defer cFree(c_output)
	if c_output == nil {
		gram_result.Err = fmt.Errorf("failed running 'C.GecoRun()' and returned a null pointer")
		return gram_result
	}
	gram_result.CorrectText = C.GoString(c_output) // Convert the C char* to a Go string
	print.Info("GEC Result: %q", gram_result.CorrectText)

	// GIBBERISH RUN
	if DoGibb {
		gibbTime := time.Now()
		// Get gibberish texts list
		gibb_texts, gibb_scores := GibbTextList(text, all_texts)
		print.Debug("Gibb Sizes: %d - %d", len(gibb_texts), len(gibb_scores))

		// Convert Go strings to C strings
		c_gibbs, cgibb_cleanup := goStringsToC(gibb_texts)
		defer cgibb_cleanup()

		// Allocate C memory for probs
		probs := make([][nClasses]float64, maxBatch)
		cProbs := (*[nClasses]C.double)(unsafe.Pointer(&probs[0])) // Type: *[4]main._Ctype_double

		// Get Gibberish Scores
		C.GecoGibb(*geco, cProbs, &c_gibbs[0], C.int(len(gibb_texts)))
		print.Info("Completed C.GecoGibb run!")

		// Convert double** to []GibbResults
		for i, gibb := range gibb_scores {
			print.Debug("Gibb[%d] running", i)
			if i >= maxBatch {
				break // safety guard – C only wrote maxBatch rows
			}
			// Skip newline literal(s) strings
			if gibb.Index == -1 {
				print.Debug("Gibb[%d] == Newline Literal String", i)
				continue
			}
			print.Debug("Gibb[%d] appending this gibb score: %+v", i, gibb)
			// Create a Go slice for the ith row
			row := probs[i] // row is [4]float64, directly readable
			gibb.Score = GibbScores{
				Clean:     float32(row[0]),
				Mild:      float32(row[1]),
				WordSalad: float32(row[2]),
				Noise:     float32(row[3]),
			}
			gram_result.GibbScores = append(gram_result.GibbScores, gibb)
		}

		print.Info("Full Gibberish Time: %v", time.Since(gibbTime).Seconds())
	}
	duration := time.Since(chanTime).Seconds()
	gram_result.ServiceTime = duration
	return gram_result
}

// Returns list of strings with original texts, sentence pairs, and full paragraphs (IN THIS ORDER)
func GibbTextList(baseText string, allTexts []string) ([]string, []GibbResults) {
	gibbScores := []GibbResults{} // Track index & length of each text
	newTexts := []string{}

	paragraph_texts := []string{}
	searchStart := 0
	pgIndex := 0 // Track index where paragraph began
	for i, text := range allTexts {
		textLen := utf8.RuneCountInString(text)
		// Newline Literal(s) String - Append string and continue
		if strings.Contains(text, "\n") {
			newTexts = append(newTexts, text)
			gibbScores = append(gibbScores, GibbResults{Index: -1, Length: -1, Score: GibbScores{}})
			searchStart += textLen
			continue
		}

		// Get index where text is found in remaining string
		text_index := RuneIndex(baseText[searchStart:], text)

		if text_index == -1 && strings.HasPrefix(text, "Summarize") {
			// Handle indexing assuming "summarize" prefix got changed to "Summarize"
			print.Debug("Indexing for text with T5 prefix")
			text_index = RuneIndex(baseText[searchStart:], text[1:]) - 1
			print.Debug("New Text Index: %v", text_index)
		}

		// Add single sentence
		newTexts = append(newTexts, text)
		gibbScores = append(gibbScores, GibbResults{Index: (text_index + searchStart), Length: textLen, Score: GibbScores{}})
		paragraph_texts = append(paragraph_texts, text)
		if len(paragraph_texts) == 1 {
			// If this was the start of the paragraph, mark its starting index
			pgIndex = text_index + searchStart
		}

		// Helper function to check for duplicates in gibbScores
		// This prevents repeated markups when adding for a paragraph or sentence pair
		hasDuplicate := func(idx, length int) bool {
			for _, g := range gibbScores {
				if g.Index == idx && g.Length == length {
					return true
				}
			}
			return false
		}

		// If the next sentence is the last OR a Newline String then complete and add the paragraph
		if (i == len(allTexts)-1) || (strings.Contains(allTexts[i+1], "\n")) {
			// Complete and append paragraph
			full_paragraph := strings.Join(paragraph_texts, " ")
			textLen = utf8.RuneCountInString(full_paragraph)
			if !hasDuplicate(pgIndex, textLen) {
				newTexts = append(newTexts, full_paragraph)
				gibbScores = append(gibbScores, GibbResults{Index: pgIndex, Length: textLen, Score: GibbScores{}})
			}
			paragraph_texts = []string{}
		} else {
			// Add sentence pair
			textPair := text + " " + allTexts[i+1]
			textLen = utf8.RuneCountInString(textPair)
			if !hasDuplicate(text_index+searchStart, textLen) {
				newTexts = append(newTexts, textPair)
				gibbScores = append(gibbScores, GibbResults{Index: (text_index + searchStart), Length: textLen, Score: GibbScores{}})
			}
		}
		searchStart += text_index
	}

	return newTexts, gibbScores
}

// Converts go strings to C strings and returns a cleanup function
func goStringsToC(strings []string) ([]*C.char, func()) {
	cstrs := make([]*C.char, len(strings))
	for i, s := range strings {
		cstrs[i] = C.CString(s)
	}

	// Return a cleanup function
	cleanup := func() {
		print.Info("Cleaning C strings")
		for _, cstr := range cstrs {
			cFree(cstr)
		}
	}
	return cstrs, cleanup
}

// Frees memory that was allocated by C.CString / C.malloc.
// The conversion to unsafe.Pointer is required by the C API and is audited.
// #nosec G103 -- audited: memory comes from C.CString, freed with C.free
func cFree(p *C.char) {
	// The #nosec tells static‑analysis tools that you did audit it.
	if p != nil {
		C.free(unsafe.Pointer(p))
	}
}
