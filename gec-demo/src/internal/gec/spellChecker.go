// src/internal/gec/spellChecker.go
package gec

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
	"embed"

	hunspell "github.com/sthorne/go-hunspell"
)

var (
	huns           *hunspell.Hunhandle
	DirtyPath     string // List of inappropriate words
	ProfanePath   string // List of words to be marked as profanity

	// Regex Patterns
	validStr = regexp.MustCompile(`^[a-zA-Z- ]+$`)  // Matches strings only made up of letters, spaces, and hyphens
	emojiRe  = regexp.MustCompile(`[\p{So}\p{Sk}]`) // Emojis
)

//go:embed data/*
var dataFS embed.FS

const (
	affixFile   = "data/index.aff"
	dictFile    = "data/index.dic"
	customFile  = "data/spelling_custom.txt"
	dirtyFile   = "data/dirty-words.txt"
	profaneFile = "data/profane-words.txt"
)

func writeTempFileFromEmbed(embedPath, prefix string) (string, error) {
	data, err := dataFS.ReadFile(embedPath)
	if err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		return "", err
	}

	return tmp.Name(), nil
}

func InitSpellChecker() (err error) {
	// Path to Affix for misspelled words
	affPath, err := writeTempFileFromEmbed(affixFile, "gec-aff-")
	if err != nil {
		return err
	}

	// Path to Dictionary for misspelled words
	dictPath, err := writeTempFileFromEmbed(dictFile, "gec-dic-")
	if err != nil {
		return err
	}

	// Files holds list of words to add to our current dictionary
	customSpellPath, err := writeTempFileFromEmbed(customFile, "gec-custom-")
	if err != nil {
		return err
	}

	DirtyPath, err = writeTempFileFromEmbed(dirtyFile, "gec-dirty-")
	if err != nil {
		return err
	}

	ProfanePath, err = writeTempFileFromEmbed(profaneFile, "gec-profane-")
	if err != nil {
		return err
	}

	// Load Hunspell using temp files
	huns = hunspell.Hunspell(affPath, dictPath)

	// Add more words to the loaded dictionary
	return addToDictionary(customSpellPath)
}

// Read in a file and add valid words to the dictionary
func addToDictionary(customSpellPath string) error {
	// Read in a file
	file, err := os.Open(customSpellPath)
	if err != nil {
		return fmt.Errorf("failed opening dictionary file: %w", err)
	}
	defer file.Close()

	// Read each line from the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())

		// Ignore blank lines or those starting with `#`
		if strings.HasPrefix(word, "#") || word == "" {
			continue
		}

		// If string is valid add it to the dictionary
		if validStr.MatchString(word) {
			huns.Add(word)
		}
	}

	// Check for errors during scanning (not including EOF)
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed scanning dictionary file: %w", err)
	}
	return nil
}

// Function to check for index collision
func checkCollision(Misspells []Misspell, newIndex, newLength int) bool {
	if IgnoreCollisions {
		// If we are ignoring collisions then always mark it as never having a collision
		return false
	}
	for _, m := range Misspells {
		// Calculate the range of existing and new misspellings
		existingStart := m.Index
		existingEnd := m.Index + m.Length - 1
		newStart := newIndex
		newEnd := newIndex + newLength - 1

		// Check if the new index range intersects with the existing range
		if (newStart <= existingEnd && newStart >= existingStart) || (newEnd >= existingStart && newEnd <= existingEnd) {
			return true // Collision detected
		}
	}
	return false // No collision
}

// Find index of substr in str (include unicodes in length)
func findRuneIndex(str, substr string, startInd int) int {
	runeIndex := 0
	byteIndex := 0

	for i, r := range str {
		// Check if the substring starts at this rune
		if len(str[i:]) >= len(substr) && str[i:i+len(substr)] == substr && runeIndex >= startInd {
			return runeIndex
		}

		// Move to the next rune
		runeIndex++
		byteIndex += utf8.RuneLen(r)
	}

	return -1
}

// Removes punctuation from the word & Removes any word containing digits or undesired characters
func cleanWord(word string) string {
	undesired := "*.,!?\"'()[]{}:;#&+-/=$%<>@_|~"

	// Format the word
	word = strings.Trim(word, undesired) // Remove leading and trailing undesired characters
	word = strings.TrimSpace(word)       // Remove an leading and trailing spaces

	// Discard word with undesired characters
	xs := []string{"\a", "\b", "\f", "\n", "\r", "\t", "\v", "\\", "\"", "-", ".", "=", "'", "`", "?", "$", "’", "“", "‘", "–"}
	for _, v := range xs {
		if strings.Contains(word, v) {
			return ""
		}
	}

	// Discard word with numeric digits or unicode characters
	for _, r := range word {
		if unicode.IsDigit(r) || r > unicode.MaxASCII {
			return ""
		}
	}

	return word
}

// SpellChecker
func SpellChecker(misspells []Misspell, data string) []Misspell {
	misspells = MarkEmojis(misspells, data)
	wordsInFile := strings.Fields(data)
	wordStartIndex := 0

	// Loop through list of words in the data
	for _, word := range wordsInFile {
		// Find the index of the current word in the data
		index := findRuneIndex(data, word, wordStartIndex)
		wordStartIndex = index + utf8.RuneCountInString(word)

		// Clean the word
		cleaned := cleanWord(word)
		cleanLen := utf8.RuneCountInString(cleaned)

		if !(huns.Spell(cleaned)) {
			// Add length of the removed prefix to the index
			index += utf8.RuneCountInString(strings.Split(word, cleaned)[0])
			suggested := huns.Suggest(cleaned)

			// Check for collisions
			if !checkCollision(misspells, index, cleanLen) {
				misspells = append(misspells, Misspell{Index: index, Length: cleanLen, Category: "SPELLING_MISTAKE", Suggestions: suggested})
			}
		}
	}

	return misspells
}

// Mark emotoicons as misspelling errors
func MarkEmojis(misspells []Misspell, text string) []Misspell {
	// Find all matches and their positions
	matches := emojiRe.FindAllStringIndex(text, -1)

	for _, match := range matches {
		// Extract emoji using the found index range
		emoji := text[match[0]:match[1]]
		idx := utf8.RuneCountInString(text[:match[0]])
		ln := utf8.RuneCountInString(emoji)
		fmt.Printf("Emoji: %q found at index: %d (Len: %d)\n", emoji, idx, ln)

		// Check for collisions
		if !checkCollision(misspells, idx, ln) {
			misspells = append(misspells, Misspell{Index: idx, Length: ln, Category: "SPELLING_MISTAKE", Suggestions: []string{}})
		}
	}
	return misspells
}

func DirtySpellChecker(data string) ([]Misspell, error) {
	// Reset Misspells to be empty
	var misspells []Misspell

	// Open and Read the "dirty-words.txt" file
	file, err := os.Open(DirtyPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening 'dirty-words.txt' file: %w", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Read the "dirty-words.txt" file line by line
	for scanner.Scan() {
		content := scanner.Text()

		// Create the regex pattern (case insensitive)
		pattern := "(?i)\\b" + string(content) + "\\b"

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed compiling regex pattern: %w", err)
		}

		// Find all occurrences of the word in the data string
		matches := re.FindAllStringIndex(data, -1)
		for _, match := range matches {
			// M[0] is the start index, M[1] is the end index
			ind := match[0]
			ln := (match[1] - match[0])

			// Check for collisions
			if !checkCollision(misspells, ind, ln) {
				misspells = append(misspells, Misspell{Index: ind, Length: ln, Category: "PROFANITY", Suggestions: nil})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed scanning file: %w", err)
	}

	// Check other bad words .txt file
	profane_file, err := os.Open(ProfanePath)
	if err != nil {
		return nil, fmt.Errorf("failed opening 'profane-words.txt' file: %w", err)
	}
	defer profane_file.Close()

	scanner = bufio.NewScanner(profane_file)
	for scanner.Scan() {
		badWord := strings.TrimSpace(scanner.Text())
		badWord = strings.ToLower(badWord)

		// Create a regex pattern with word boundaries
		pattern := fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(badWord))

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed compiling profane regex pattern: %w", err)
		}

		// Find all occurrences of the word in the data string
		matches := re.FindAllStringIndex(data, -1)
		for _, match := range matches {
			// M[0] is the start index, M[1] is the end index
			ind := match[0]
			ln := (match[1] - match[0])

			// Check for collisions
			if !checkCollision(misspells, ind, ln) {
				misspells = append(misspells, Misspell{Index: ind, Length: ln, Category: "PROFANITY", Suggestions: nil})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed scanning profane file: %w", err)
	}

	return misspells, nil
}
