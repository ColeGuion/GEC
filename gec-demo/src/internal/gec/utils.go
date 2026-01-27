// src/internal/gec/utils.go
package gec

import (
	"fmt"
	"gec-demo/src/internal/print"
	"gec-demo/src/internal/speechtagger"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Formats the `Differences` & `Misspells` as JSON text markups return
func FormatToJson(text string, Differences []Markup, Misspells []Misspell) (markups []Markup, err_chars int, profanity_words []string, err error) {
	text_length := utf8.RuneCountInString(text)
	err_chars = 0

	// Iterate over 'Differences' and create JSON for each
	for _, diff := range Differences {
		err_chars += diff.Length
		
		// Skip markups covering only newline literals and return chars
		substr, err := GetSubstring(text, diff.Index, diff.Length)
		if err != nil {
			print.Error("GetSubstring() failed for markup, %v", err)
			continue
		}
		if strings.Trim(substr, "\r\n") == "" {
			continue
		}

		// Skip markups which index beyond the size of the original text
		if (diff.Length + diff.Index) > text_length {
			print.Warning("Markup '%v' indexes beyond the original text. Text Size: %v. Diff Markup: %+v", diff.Category, text_length, diff)
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
			print.Error("GetSubstring() failed for markup, %v", err)
			continue
		}
		if strings.TrimSpace(substr) == "" {
			continue
		}

		// Skip markups which index beyond the size of the original text
		if (miss.Length + miss.Index) > text_length {
			print.Warning("Markup '%v' indexes beyond the original text. Text Size: %v. Misspelling Markup: %+v", miss.Category, text_length, miss)
			continue
		}

		if miss.Category == "SPELLING_MISTAKE" {
			typo := Markup{
				Index:    miss.Index,
				Length:   miss.Length,
				Message:  "Possible spelling mistake found.",
				Category: miss.Category,
			}
			markups = append(markups, typo)
		}
		if miss.Category == "PROFANITY" {
			dirtyMark := Markup{
				Index:    miss.Index,
				Length:   miss.Length,
				Message:  "This word is considered offensive",
				Category: miss.Category,
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

	// Sort by Offsets
	sort.Slice(markups, func(i, j int) bool {
		return markups[i].Index < markups[j].Index
	})
	print.Debug("Text Markups: %+v\n", markups)

	if err_chars > 0 && len(markups) == 0 {
		print.Warning("Error Character Count(%d) > 0 && Text markups is empty", err_chars)
	}

	return markups, err_chars, profanity_words, err
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
	for i := range allTexts {
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
		return "", fmt.Errorf("invalid start index(%v) or length(%v). Input string length=%v", startInd, length, len(str))
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
