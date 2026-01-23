// src/internal/gec/gec.go
package gec

/*
#cgo CFLAGS: -I${SRCDIR}/../../native/gec_runtime/include
#cgo CFLAGS: -I${SRCDIR}/../../native/gec_runtime/third_party
#cgo CFLAGS: -I${SRCDIR}/../../native/gec_runtime/third_party/onnxruntime/include
#cgo CFLAGS: -I${SRCDIR}/../../native/gec_runtime/third_party/sentencepiece/include

#cgo LDFLAGS: ${SRCDIR}/../../native/gec_runtime/build/libgec.a
#cgo LDFLAGS: -L${SRCDIR}/../../native/gec_runtime/third_party/onnxruntime/lib
#cgo LDFLAGS: -L${SRCDIR}/../../native/gec_runtime/third_party/sentencepiece/lib
#cgo LDFLAGS: -L${SRCDIR}/../../native/gec_runtime/third_party/icu/lib
#cgo LDFLAGS: -lonnxruntime -lsentencepiece -lstdc++ -lm -ldl -ljson-c -licuuc -licudata -lonnxruntime_providers_shared -lonnxruntime_providers_cuda

#include "inference.h"
*/
import "C"
import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unsafe"

	"gec-demo/src/internal/print"
	"gec-demo/src/internal/speechtagger"
)

var (
	rePrefix  = regexp.MustCompile(`(?i)^translate English to (german|french|romanian)`) // Regex to match "Translate English to (German|French|Romanian)" case-insensitively
	rePreproc = regexp.MustCompile(`\s*\n+\s*`)

	CountLT        = 0
	ChanCapacity   = 250
	DeviceCount    = 1
	NumGpuChannels = 1
	GecoChannels   []chan WorkItem

	IgnoreCollisions = false
	DoMisspellings   = true
)

func init() {
	print.SetLevel(0) // DEBUG
	GecoChannels = make([]chan WorkItem, NumGpuChannels)
	print.Debug("Total GEC Channels: %d", NumGpuChannels)

	//TODO: Set from config or elsewhere
	cLogLevel := 4

	for i := range NumGpuChannels {
		gpuId := i % DeviceCount
		print.Debug("Creating Geco[%d] for gpu-%v", i, gpuId)

		// Buffered channel with a capacity of `ChanCapacity`
		GecoChannels[i] = make(chan WorkItem, ChanCapacity)
		go ClaimGpu(cLogLevel, gpuId, GecoChannels[i])
	}

	// Initialize the parts-of-speech tagging model
	err := speechtagger.InitTaggingModel()
	if err != nil {
		fmt.Printf("ERROR: Failed to initialize TaggerModel: %v\n", err)
		return
	}

	if DoMisspellings {
		err = InitSpellChecker()
		if err != nil {
			print.Error("failed to initialize SpellChecker: %v\n", err)
			return
		}
	}
}

func ClaimGpu(logLevel int, gpuId int, ch chan WorkItem) {
	var geco unsafe.Pointer

	// Allocate a Geco object for the channel
	geco = C.NewGeco(C.int(logLevel), 0, C.int(gpuId))
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
	text_markups, err_chars, profanity_words, err := FormatToJson(text, differences, misspells)
	if err != nil {
		return nil, fmt.Errorf("error in FormatToJson(), %w", err)
	}

	gec_result.CorrectedText = corrected_text
	gec_result.TextMarkups = text_markups
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

	duration := time.Since(chanTime).Seconds()
	gram_result.ServiceTime = duration
	return gram_result
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
