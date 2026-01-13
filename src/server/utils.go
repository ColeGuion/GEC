// utils.go
package main

import (
	_"encoding/json"

	"fmt"
	"gec-api/print"
	"gec-api/speechtagger"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Formats the `Differences` & `Misspells` as JSON text markups return
// TODO: Fix error handling
func FormatToJson(text string, Differences []Markup, Misspells []Misspell, gibberishScores []GibbResults) (markups []Markup, err_chars int, profanity_words []string, err error) {
	//var markups []Markup
	var gibberish_texts []GibbResults
	//var profanity_words []string
	text_length := utf8.RuneCountInString(text)
	err_chars = 0

	// Filter out objects so you are left with those that don't pass the gibberish scores thresholds
	for _, gibb := range gibberishScores {
		// Too small of a Clean score or too large of another score (Mild, Noise, Salad) will add you to the gibberish_texts array
		if (int(gibb.Score.Clean) < CleanThreshold) && (int(gibb.Score.Mild) > MildThreshold || int(gibb.Score.Noise) > NoiseThreshold || int(gibb.Score.WordSalad) > SaladThreshold) {
			gibberish_texts = append(gibberish_texts, gibb)
		} else {
			// Safely extract the text with bounds checking
			substr, err := GetSubstring(text, gibb.Index, gibb.Length)
			if err != nil {
				fmt.Printf("error in GetSubstring() for gibberish check: %v\n", err)
				continue
			}

			// Mark as gibberish if there are 6-single letter words for every normal word in a string
			ratio := singleLetterRatio(substr)
			if ratio > 6.0 {
				gibberish_texts = append(gibberish_texts, gibb)
			}
		}
	}

	// Track total error character count
	for _, score := range gibberish_texts {
		err_chars += score.Length
	}

	// Iterate over 'Differences' and create JSON for each
	for _, diff := range Differences {
		err_chars += diff.Length
		// Skip markups covering only newline literals and return chars
		substr, err := GetSubstring(text, diff.Index, diff.Length)
		if err != nil {
			fmt.Printf("error in GetSubstring() for markup, %v", err)
			continue
		}
		if strings.Trim(substr, "\r\n") == "" {
			continue
		}

		// Keep gibberish texts that don't intersect with diffs
		gibberish_texts = filterIntersectingGibberish(diff.Index, diff.Length, gibberish_texts)

		// Skip markups which index beyond the size of the original text
		if (diff.Length + diff.Index) > text_length {
			print.Warning("Markup '%v' indexes beyond the original text. Text Size: %v. Diff Markup: %+v", diff.Type, text_length, diff)
			continue
		}

		markups = append(markups, diff)
	}

	// Iterate over 'Misspells' and create JSON for each
	for _, miss := range Misspells {
		err_chars += miss.Length

		// Skip misspelling markups marking nothing
		substr, err := GetSubstring(text, miss.Index, miss.Length)
		if err != nil {
			fmt.Printf("error in GetSubstring() for markup, %v", err)
			continue
		}
		if strings.TrimSpace(substr) == "" {
			continue
		}

		// Keep gibberish texts that don't intersect with misspellings
		gibberish_texts = filterIntersectingGibberish(miss.Index, miss.Length, gibberish_texts)

		// Skip markups which index beyond the size of the original text
		if (miss.Length + miss.Index) > text_length {
			print.Warning("Markup '%v' indexes beyond the original text. Text Size: %v. Misspelling Markup: %+v", miss.Type, text_length, miss)
			continue
		}

		if miss.Type == "typo" {
			//if miss.Type == "SPELLING_MISTAKE" {
			typo := Markup{
				Index:   miss.Index,
				Length:  miss.Length,
				Message: "Possible spelling mistake found.",
				Type:    miss.Type,
			}
			markups = append(markups, typo)
		}
		if miss.Type == "dirty" {
			//if miss.Type == "PROFANITY" {
			dirtyMark := Markup{
				Index:   miss.Index,
				Length:  miss.Length,
				Message: "This word is considered offensive",
				Type:    miss.Type,
			}
			markups = append(markups, dirtyMark)

			// Add bad word to profanity list
			badWord, err := GetSubstring(text, miss.Index, miss.Length)
			if err != nil {
				profanity_words = append(profanity_words, fmt.Sprintf("Invalid Markup: {Index: %d, Length: %d}", miss.Index, miss.Length))
			} else {
				profanity_words = append(profanity_words, badWord)
			}
		}
	}

	// Loop through gibberish_texts
	for len(gibberish_texts) > 0 {
		mark_type := "GIBBERISH"
		gibb := gibberish_texts[0]
		gibberish_texts = gibberish_texts[1:]

		// Remove intersecting gibberish values
		// Should set the priority of: Sentences > Sentence Pairs > Paragraphs
		gibberish_texts = filterIntersectingGibberish(gibb.Index, gibb.Length, gibberish_texts)

		// Skip markups which index beyond the size of the original text
		if (gibb.Length + gibb.Index) > text_length {
			print.Warning("Markup '%v' indexes beyond the original text. Text Size: %v. Gibb Markup: %+v", mark_type, text_length, gibb)
			continue
		}

		// Marked up gibberish sentence HERE
		/* substr, err := GetSubstring(text, gibb.Index, gibb.Length)
		if err != nil {
			fmt.Printf("error in GetSubstring(%v, %v, %q) for markup, %v", gibb.Index, gibb.Length, text, err)
		} */

		gibb_markup := Markup{
			Index:   gibb.Index,
			Length:  gibb.Length,
			Message: "Text is unclear.",
			Type:    mark_type,
		}
		markups = append(markups, gibb_markup)
	}

	// Sort by Offsets
	sort.Slice(markups, func(i, j int) bool {
		return markups[i].Index < markups[j].Index
	})
	print.Debug("Text Markups: %+v\n", markups)

	if err_chars > 0 && len(markups) == 0 {
		print.Warning("Error Character Count(%d) > 0 && Text markups is empty", err_chars)
	}

	
	return markups, err_chars, profanity_words, err
	/* // Wrap the markup errors and convert to a []byte
	wrapper := Wrapper{
		TextMarkups: markups,
		ErrorCharacters: err_chars,
		ProfaneWords: profanity_words,
	}

	// Convert the markup errors to a []byte
	jsonData, err = json.MarshalIndent(wrapper, "", "  ")
	return jsonData, err */
}

// Split the text into sentences and newline literals with surrounding whitespace
func PreprocessText(text string) (allTexts []string) {
	text = CleanText(text)
	inds := rePreproc.FindAllStringIndex(text, -1)

	lastIndex := 0
	for _, ind := range inds {
		start, end := ind[0], ind[1]
		if lastIndex < start {
			// Split text into sentences and append
			sents := speechtagger.SplitBySentences(text[lastIndex:start])
			for _, s := range sents {
				if s != "" {
					allTexts = append(allTexts, s)
				}
			}
		}

		// Trim the beginning of the newline literal before the newline
		lines := strings.Split(text[start:end], "\n")
		lines[0] = strings.TrimSpace(lines[0])
		newlineStr := strings.Join(lines, "\n")

		// Append newline whitespace literals
		allTexts = append(allTexts, newlineStr)
		lastIndex = end
	}
	if lastIndex < len(text) {
		// Split text into sentences and append
		sents := speechtagger.SplitBySentences(text[lastIndex:])

		for _, s := range sents {
			if s != "" {
				allTexts = append(allTexts, s)
			}
		}
	}

	// Modify text starting with a T5 prefix
	for i, _ := range allTexts {
		if strings.HasPrefix(allTexts[i], "summarize") {
			allTexts[i] = strings.Replace(allTexts[i], "summarize", "Summarize", 1)
		}

		// Convert everything after "Translate" to lowercase, if the string matches this regex prefix
		allTexts[i] = rePrefix.ReplaceAllStringFunc(allTexts[i], func(match string) string {
			// Extract the matched part after "Translate"
			parts := rePrefix.FindStringSubmatch(match)
			if len(parts) == 2 {
				return fmt.Sprintf("Translate english to %s", strings.ToLower(parts[1]))
			}
			return match
		})
	}

	return allTexts
}

// Clean the Text of weird characters
func CleanText(text string) string {
	// Strip out any control characters that are not printable
	ru := []rune(text)
	cleanData := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			print.Debug("Control Character being dropped: %v", r)
			return -1
		}
		if r == 160 {
			return 32
		}
		return r
	}, string(ru))

	// Filter special quote characters (“, ”, ’, ‘)
	cleanData = strings.ReplaceAll(cleanData, "“", "\"")
	cleanData = strings.ReplaceAll(cleanData, "”", "\"")
	cleanData = strings.ReplaceAll(cleanData, "‘", "'")
	cleanData = strings.ReplaceAll(cleanData, "’", "'")
	return cleanData
}

// Returns the substring of a string given the starting index and length of the substring
func GetSubstring(str string, startInd int, length int) (string, error) {
	if startInd == len(str) && length == 1 {
		// Wants to markup end of string
		return "", nil
	}
	if startInd < 0 || length <= 0 || startInd+length > len(str) {
		return "", fmt.Errorf("invalid start index(%v) or length(%v)", startInd, length)
	}
	return str[startInd : startInd+length], nil
}

// Get leading & trailing whitespace around a string
func getSpaceAround(s string) (string, string) {
	var begSpace, trailingWhitespace strings.Builder

	// Find leading whitespace
	i := 0
	for i < len(s) && unicode.IsSpace(rune(s[i])) {
		begSpace.WriteByte(s[i])
		i++
	}

	// Find trailing whitespace
	j := len(s) - 1
	for j >= i && unicode.IsSpace(rune(s[j])) {
		trailingWhitespace.WriteByte(s[j])
		j--
	}

	// Reverse trailingWhitespace because we collected it from the end of the string
	chars := []rune(trailingWhitespace.String())
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	endSpace := string(chars)

	return begSpace.String(), endSpace
}

// Checks if a slice contains a specific element.
func contains(slice []string, item string) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

// Checks if two markups will collide
func intersects(index1, length1, index2, length2 int) bool {
	end1 := index1 + length1
	end2 := index2 + length2

	// Check if ranges overlap
	return index1 < end2 && index2 < end1
}

// Filters out gibberish texts that intersect with the given index and length
func filterIntersectingGibberish(index, length int, gibberish_texts []GibbResults) []GibbResults {
	if IgnoreCollisions {
		// If we are not searching for intersecting markups then return the passed in array
		return gibberish_texts
	}
	var newGibbs []GibbResults
	for _, gibb := range gibberish_texts {
		if !intersects(index, length, gibb.Index, gibb.Length) {
			newGibbs = append(newGibbs, gibb)
		}
	}
	return newGibbs
}

func singleLetterRatio(input string) float64 {
	var singleLetterCount, multiLetterCount int
	// Split the string into words
	words := strings.Fields(input)

	for _, word := range words {
		// Remove punctuation
		cleaned := strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r)
		})

		if len(cleaned) == 1 {
			singleLetterCount++
		} else if len(cleaned) > 1 {
			multiLetterCount++
		}
	}

	// Get the ratio
	if multiLetterCount == 0 {
		return float64(singleLetterCount)
	} else {
		return float64(singleLetterCount) / float64(multiLetterCount)
	}
}

// Get accurate index of substring in a string, account for runes
func RuneIndex(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	// Convert to rune slices to work with characters
	runes := []rune(s)
	subrunes := []rune(substr)

	for i := 0; i <= len(runes)-len(subrunes); i++ {
		if string(runes[i:i+len(subrunes)]) == substr {
			return i
		}
	}
	return -1
}
