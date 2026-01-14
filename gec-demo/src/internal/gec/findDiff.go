// src/internal/gec/findDiff.go
package gec

import (
	"bytes"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"gec-demo/src/internal/print"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var (
	//reCase1     = regexp.MustCompile(`\x1b\[31m(\s*)\x1b\[0m(\s*\n+)`)
	reCase2     = regexp.MustCompile(`([.?!][ ]?)\x1b\[31m([ ]*)\x1b\[0m`)
	reCase3     = regexp.MustCompile(`[[:alnum:],]*((\x1b\[3[12]m.*?\x1b\[0m)(\x1b\[3[12]m.*?\x1b\[0m|\S)*)`)
	reCaseQuote = regexp.MustCompile(`[[:alnum:],.?!]*((\x1b\[3[12]m"*?\x1b\[0m)(\x1b\[3[12]m"*?\x1b\[0m|\S)*)`)
	subMatch1   = regexp.MustCompile(`\x1b\[31m\s`)
	subMatch2   = regexp.MustCompile(`\x1b\[31m.*?\s\x1b\[0m`)
	anyChangeRe = regexp.MustCompile(`\x1b\[(3[12])m(.*?)\x1b\[0m`) // Find any changes in the text
	goodEscapes = regexp.MustCompile(`\x1b\[32m(.*?)\x1b\[0m`)      // Matches good escape sequences and its contents
	badEscapes  = regexp.MustCompile(`\x1b\[31m(.*?)\x1b\[0m`)      // Matches bad escape sequences and its contents
	allEscapes  = regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)           // Matches all escape sequences but NOT their contents
	swapRe      = regexp.MustCompile(`(\x1b\[31m)(\s*)`)            // Matches bad escape sequence followed by whitespace
)

func FindDifference(t1, t2 string, Misspells []Misspell) ([]Markup, error) {
	var Differences []Markup // Tracks differences in text
	myStr := ""              // Running string to keep track of the buffer string

	// Split texts into lines
	linesOne := strings.Split(t1, "\n")
	linesTwo := strings.Split(t2, "\n")
	if len(linesOne) > len(linesTwo) {
		print.Warning("FindDiff() Error: Original and Connected text split into lines of uneven in length (%v vs %v)", len(linesOne), len(linesTwo))
	}

	for i, l1 := range linesOne {
		var l2 string
		if i+1 > len(linesTwo) {
			l2 = linesOne[i] // Make it the same as original text
			//return nil, fmt.Errorf("FindDiff() Error: Original and Connected text split into lines of uneven in length (%v vs %v)", len(linesOne), len(linesTwo))
		} else {
			l2 = linesTwo[i]
		}

		if l1 != l2 {
			// Run Diff Finder on the string with the original line and the corrected line
			DiffFinder((myStr + l1), (myStr + l2), Misspells, &Differences)
		}
		// Append a newline and the original text
		myStr += l1 + "\n"
	}

	// Sort the diffs slice based on the Index field
	sort.Slice(Differences, func(i, j int) bool {
		return Differences[i].Index < Differences[j].Index
	})

	return Differences, nil
}

// Find and mark up any Differences between the strings
func DiffFinder(t1, t2 string, Misspells []Misspell, Differences *[]Markup) {
	// Setting Differences back to an empty array
	var matches []string
	var allMatches [][]string

	// Get the buffer string
	buffStr := getBuffer(t1, t2)

	// CASE #2: Ignore removed space before a new sentence
	for reCase2.MatchString(buffStr) {
		matches = reCase2.FindStringSubmatch(buffStr)
		buffStr = strings.Replace(buffStr, matches[0], (matches[1] + matches[2]), -1)
	}

	// CASE QUOTE: Like case 3 but just for quotes
	skipCases := []string{}
	for reCaseQuote.MatchString(buffStr) {
		allMatches = reCaseQuote.FindAllStringSubmatch(buffStr, -1)

		// If total matches equals the number of skip cases, then break
		if len(allMatches) == len(skipCases) {
			break
		}

		for _, matches := range allMatches {
			// Skip case if its in the skipCases array
			if contains(skipCases, matches[0]) {
				continue
			}

			// A markup is cut off in our match
			if strings.LastIndex(matches[0], "\x1b[3") > strings.LastIndex(matches[0], "\x1b[0") {
				skipCases = append(skipCases, matches[0])
				continue
			}

			matches[0] = strings.TrimPrefix(matches[0], "0m")
			runCase(Differences, &buffStr, matches[0], Misspells)
		}
	}

	// CASE #3: All encompassing case. Matches all text that include a modifications
	// Start with Letters, Numbers, or Commas
	for reCase3.MatchString(buffStr) {
		// Loop through until all matches are resolved
		allMatches = reCase3.FindAllStringSubmatch(buffStr, -1)

		for _, matches := range allMatches {
			switch {
			case matches[0] == "\x1b[32m \x1b[0m\"":
				//! May be able to remove as this maybe a vennify specific problem
				// Ignore added spaces before ending quotes
				// Do this by replacing the 1st instance of Matches[0] with just the quote `"`
				buffStr = strings.Replace(buffStr, matches[0], matches[3], 1)

			case subMatch1.MatchString(matches[1]) && matches[1] == matches[2]:
				// First modification is removal with whitespace
				// Put the white space before escape literal and then run the case
				swapped := swapSpace(matches[1])
				buffStr = strings.Replace(buffStr, matches[1], swapped, 1)
				runCase(Differences, &buffStr, strings.TrimSpace(swapped), Misspells)

			case matches[1] == matches[0] && subMatch2.MatchString(matches[2]):
				// Removal before normal unrelated text
				runCase(Differences, &buffStr, matches[2], Misspells)

			default:
				runCase(Differences, &buffStr, matches[0], Misspells)
			}
		}
	}
}

// Returns value to add to []Markup slice
func runCase(diffs *[]Markup, buffStr *string, match string, Misspells []Misspell) {
	var replMsg, msgType string
	ind := getWordsIndex(*buffStr, match)

	// Remove unmodified punctuation marks from the end of the match
	newMatch := strings.TrimRight(match, ".,?!:;")

	origWord, replWord := getWords(newMatch)

	// Get length of word in runes
	wordLen := utf8.RuneCountInString(origWord)

	// Modify buffStr by removing this occurence with the original word
	*buffStr = strings.Replace(*buffStr, newMatch, origWord, 1)

	// Ignore replacements over 40 times the length of the original word
	if len(replWord) > len(origWord)*40 && len(origWord) > 1 {
		return
	}

	// Get replacement message and add to diffs
	msgType, replMsg = getMsg(newMatch, replWord, origWord)
	addToDiffs(diffs, ind, wordLen, replMsg, msgType, Misspells)
}

// Return added & removed changes in the word
func getChanges(word string) (string, string) {
	var addChange []string
	var delChange []string

	matches := anyChangeRe.FindAllStringSubmatch(word, -1)
	for _, match := range matches {
		switch match[1] {
		case "32":
			addChange = append(addChange, match[2])
		case "31":
			delChange = append(delChange, match[2])
		}
	}

	return strings.Join(addChange, ""), strings.Join(delChange, "")
}

// Get replacement message for diff
func getMsg(word, repl, orig string) (string, string) {
	addChanges, delChanges := getChanges(word)
	changes := addChanges + delChanges

	// Changes only involve whitespace
	if strings.TrimSpace(changes) == "" {
		switch {
		case addChanges == "":
			return "Grammar", "Remove spacing \u201c" + repl + "\u201d"
		case delChanges == "":
			return "Grammar", "Add spacing \u201c" + repl + "\u201d"
		default:
			// Both added and removed whitespace
			return "Grammar", "Change spacing \u201c" + repl + "\u201d"
		}
	}

	// Only contains big 4 of punctuation marks
	if strings.Trim(changes, ".,?!") == "" {
		switch {
		case addChanges != "" && delChanges != "":
			return "Grammar", "Replace \u201c" + delChanges + "\u201d with \u201c" + addChanges + "\u201d"
		case strings.Contains(delChanges, ","):
			return "Grammar", "Remove comma \u201c" + repl + "\u201d"
		case strings.Contains(addChanges, ","):
			return "Grammar", "Add comma \u201c" + repl + "\u201d"
		case strings.Contains(delChanges, "."):
			return "Grammar", "Remove period \u201c" + repl + "\u201d"
		case strings.Contains(addChanges, "."):
			return "Grammar", "Add period \u201c" + repl + "\u201d"
		case strings.Contains(delChanges, "?"):
			return "Grammar", "Remove question mark \u201c" + repl + "\u201d"
		case strings.Contains(addChanges, "?"):
			return "Grammar", "Add question mark \u201c" + repl + "\u201d"
		case strings.Contains(delChanges, "!"):
			return "Grammar", "Remove exclamation mark \u201c" + repl + "\u201d"
		case strings.Contains(addChanges, "!"):
			return "Grammar", "Add exclamation mark \u201c" + repl + "\u201d"
		}
	}

	// Changes involve only specific punctuation marks and not an empty string
	if strings.Trim(changes, ".,?!:;\"") == "" {
		switch {
		case addChanges == "" && delChanges != "" && repl == "":
			return "Grammar", "Remove unnecessary punctuation."
		default:
			return "Grammar", "Punctuation Suggestion \u201c" + repl + "\u201d"
		}
	}

	switch {
	// Deleted text
	case repl == "":
		return "Grammar", "This text is unnecessary."

	// Case insensitive match
	case strings.EqualFold(repl, orig):
		return "Grammar", "Change the capitalization \u201c" + repl + "\u201d"

	// Default Case
	default:
		return "Grammar", "Did you mean \u201c" + repl + "\u201d?"
	}
}

// Returns the buffer string
func getBuffer(t1, t2 string) string {
	// Get the buffer string
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(t1, t2, false)

	// Mark up the text with insertions & deletions
	prettyTxt := dmp.DiffPrettyText(diffs)
	buffer := bytes.NewBufferString(prettyTxt) // TYPE: bytes.buffer
	return buffer.String()
}

// Moves whitespace before the escape literal (\x1b[31m) for indexing
func swapSpace(text string) string {
	// Replace occurrences using a function
	replacedText := swapRe.ReplaceAllStringFunc(text, func(match string) string {
		matches := swapRe.FindStringSubmatch(match)
		if len(matches) >= 3 {
			// Reorder the matches to put whitespace before the escape sequence
			return matches[2] + matches[1]
		}
		return match
	})

	return replacedText
}

// Add a response to the diffs slice
func addToDiffs(diffs *[]Markup, index, length int, replacement, diffType string, Misspells []Misspell) {
	// For adding words
	if length == 0 {
		length = 1
	}

	newMarkup := Markup{
		Index:   index,
		Length:  length,
		Message: replacement,
		Category:    strings.ToUpper(diffType + "_Suggestion"),
	}

	diffStart := index
	diffEnd := index + length

	//* Remove value if intersecting with a misspelling (and we do NOT ignore collisions)
	if !IgnoreCollisions {
		for _, miss := range Misspells {
			// If ranges intersect, Return from this function, do NOT add to diffs
			if miss.Index < diffEnd && diffStart < (miss.Index+miss.Length) {
				return
			}
		}
	}

	//* Remove value if intersecting with current differences found
	for _, df := range *diffs {
		if df.Index < diffEnd && diffStart < (df.Index+df.Length) {
			return
		}
	}

	*diffs = append(*diffs, newMarkup)
}

// Get index of string inside buffer string
func getWordsIndex(inpString, pattern string) int {
	parts := strings.Split(inpString, pattern)[0]

	// Removes good escape sequences and their contents
	parts = goodEscapes.ReplaceAllString(parts, "")

	// Removes all escape sequences but NOT their contents
	parts = allEscapes.ReplaceAllString(parts, "")

	return utf8.RuneCountInString(parts)
}

// Return the original and replacement value for a word (Words before and after modifications)
func getWords(word string) (string, string) {
	// Get Original & Replacement words
	origWord := goodEscapes.ReplaceAllString(word, "")
	replWord := badEscapes.ReplaceAllString(word, "")

	// Remove the rest of the escapes
	origWord = allEscapes.ReplaceAllString(origWord, "")
	replWord = allEscapes.ReplaceAllString(replWord, "")

	return origWord, replWord
}
