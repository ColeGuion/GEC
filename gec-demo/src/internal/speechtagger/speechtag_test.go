package speechtagger

import (
	"fmt"
	"reflect"
	"testing"
)

// Print the result of a failed test
func failed_test(result []string, exp []string) string {
	failStr := fmt.Sprintf("\n(%v) Result: [ ", len(result))
	for _, r := range result {
		failStr += fmt.Sprintf("%q ", r)
	}
	failStr += fmt.Sprintf("]\n(%v) Expected: [ ", len(exp))
	for _, r := range exp {
		failStr += fmt.Sprintf("%q ", r)
	}
	return failStr + "]"
}

func TestInitTaggingModel(t *testing.T) {
	// Functions Tested: InitTaggingModel(), decodeGob()
	// Call the InitTaggingModel function
	err := InitTaggingModel()

	// Check if the initialization returned an error
	if err != nil {
		t.Errorf("InitTaggingModel() returned an error: %v", err)
	}

	// Check if the TaggerModel is initialized
	if TaggerModel == nil {
		t.Errorf("TaggerModel is nil after initialization")
	}

	// Check if the SentTokenizer is initialized
	if SentTokenizer == nil {
		t.Errorf("SentTokenizer is nil after initialization")
	}

	if TagsGob == "" || WeightsGob == "" {
		t.Errorf("ERROR: TagsGob or WeightsGob is empty")
	}
}

func TestSplitBySentences(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "Empty string",
			text:     "",
			expected: []string{},
		},
		{
			name:     "Single sentence",
			text:     "Hello world.",
			expected: []string{"Hello world."},
		},
		{
			name:     "Multiple sentences",
			text:     "Hello world. How are you?",
			expected: []string{"Hello world.", "How are you?"},
		},
		{
			name:     "Sentence with newline",
			text:     "Hello world.\nHow are you?",
			expected: []string{"Hello world.", "How are you?"},
		},
		{
			name:     "Sentence with multiple newlines",
			text:     "Hello world.\n\nHow are you?",
			expected: []string{"Hello world.", "How are you?"},
		},
		{
			name:     "Multiple Whitespace",
			text:     "Hello world.\n  \t\n\nHow are you? I am good.\n\t\t\tLet's go eat.\t\t\nI am hungry.",
			expected: []string{"Hello world.", "How are you?", "I am good.", "Let's go eat.", "I am hungry."},
		},
		{
			name:     "Sentence with numbers",
			text:     "In 2023, the world changed. What happened? I was just 23 years old and my brother was 8.",
			expected: []string{"In 2023, the world changed.", "What happened?", "I was just 23 years old and my brother was 8."},
		},
		{
			name:     "Sentence with punctuation",
			text:     "Hello, world! How are you doing today?",
			expected: []string{"Hello, world!", "How are you doing today?"},
		},
		{
			name: "Test Abbreviations",
			text: "Dr. Smith and Mrs. Johnson visited the U.S.A. last year. They stayed at a hotel on St. Patrick's Ave. and met with Prof. Brown from M.I.T. Their favorite book was 'E.T.A. Hoffmann's Tales.' Mr. Doe also joined them at 10 a.m. for a discussion on AI and NASA.",
			//* Would prefer for "Their favorite book was 'E.T.A." and "Hoffmann's Tales.'" to be combined into one sentence
			expected: []string{
				"Dr. Smith and Mrs. Johnson visited the U.S.A. last year.",
				"They stayed at a hotel on St. Patrick's Ave. and met with Prof. Brown from M.I.T.",
				"Their favorite book was 'E.T.A.",
				"Hoffmann's Tales.'",
				"Mr. Doe also joined them at 10 a.m. for a discussion on AI and NASA.",
			},
		},
		{
			name:     "Test Ellipsis",
			text:     "This is a sentence with ellipsis... Here's another one... And yet another...",
			expected: []string{"This is a sentence with ellipsis... Here's another one... And yet another..."},
		},
		{
			name:     "Multiple Punctuation",
			text:     "Where are we going???? I love shopping!! What's for dinner?!?! I'm so excited.!!.!",
			expected: []string{"Where are we going????", "I love shopping!!", "What's for dinner?!?!", "I'm so excited.!!.!"},
		},
		{
			name:     "Semicolons Test",
			text:     "Sun., Feb. 12; Thurs., Oct. 31; Wed., Dec. 9; Fri., Sept. 23, 1988",
			expected: []string{"Sun., Feb. 12; Thurs., Oct. 31; Wed., Dec. 9; Fri., Sept. 23, 1988"},
		},
		{
			name: "Tabs Test",
			text: "\tHello world \t\t How are you? \t\t I am so good\t\t.\t\t\t Let's go eat!\n\v\v I love tabs like '\v' and '\t'!\v\v\t\v",
			expected: []string{
				"Hello world \t\t How are you?",
				"I am so good\t\t.",
				"Let's go eat!",
				"I love tabs like '\v' and '\t'!",
				"",
			},
		},
		{
			name:     "Emojis",
			text:     "This is my face: (-_-) I am so happy! :D",
			expected: []string{"This is my face: (-_-) I am so happy!", ":D"},
		},
		{
			name:     "Unicode and Hexadecimal",
			text:     "I \u2764 you! The first letter is '\x41'!",
			expected: []string{"I \u2764 you!", "The first letter is '\x41'!"},
		},
		{
			name:     "Backslashes and Quotes",
			text:     "Backslash goes here \\. Double quote goes here \".",
			expected: []string{"Backslash goes here \\.", "Double quote goes here \"."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitBySentences(tt.text)
			if !reflect.DeepEqual(result, tt.expected) {
				failStr := failed_test(result, tt.expected)
				t.Errorf(failStr)
			}
		})
	}
}

func TestTagSpeech(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []*Token
	}{
		{
			name:     "Empty string",
			text:     "",
			expected: []*Token{},
		},
		{
			name: "Single Sentence",
			text: "Hello, world!",
			expected: []*Token{
				{Text: "Hello", Tag: "NNP"},
				{Text: ",", Tag: ","},
				{Text: "world", Tag: "NN"},
				{Text: "!", Tag: "."},
			},
		},
		{
			name: "Quotes",
			text: "Shari said, \"How aren't you?\" Jake replied, \"I've been great!!!\"",
			expected: []*Token{
				{Text: "Shari", Tag: "NNP"},
				{Text: "said", Tag: "VBD"},
				{Text: ",", Tag: ","},
				{Text: "\"", Tag: "NNP"},
				{Text: "How", Tag: "NNP"},
				{Text: "are", Tag: "VBP"},
				{Text: "n't", Tag: "RB"},
				{Text: "you", Tag: "PRP"},
				{Text: "?", Tag: "."},
				{Text: "\"", Tag: "''"},
				{Text: "Jake", Tag: "NNP"},
				{Text: "replied", Tag: "VBD"},
				{Text: ",", Tag: ","},
				{Text: "\"", Tag: "NNP"},
				{Text: "I've", Tag: "NNP"},
				{Text: "been", Tag: "VBN"},
				{Text: "great", Tag: "JJ"},
				{Text: "!", Tag: "."},
				{Text: "!", Tag: "."},
				{Text: "!", Tag: "."},
				{Text: "\"", Tag: "NN"},
			},
		},
		{
			name: "Test #3",
			text: "Dr. Smith and Mrs. Johnson visited the U.S.A. last year. Tabs: \v\v\t\v\n!?!! All special \x1b[1;31mCOLORED\x1b[0m chars",
			expected: []*Token{
				{Text: "Dr.", Tag: "NNP"},
				{Text: "Smith", Tag: "NNP"},
				{Text: "and", Tag: "CC"},
				{Text: "Mrs.", Tag: "NNP"},
				{Text: "Johnson", Tag: "NNP"},
				{Text: "visited", Tag: "VBD"},
				{Text: "the", Tag: "DT"},
				{Text: "U.S.A.", Tag: "NNP"},
				{Text: "last", Tag: "JJ"},
				{Text: "year", Tag: "NN"},
				{Text: ".", Tag: "."},
				{Text: "Tabs", Tag: "NN"},
				{Text: ":", Tag: ":"},
				{Text: "!", Tag: "."},
				{Text: "?", Tag: "."},
				{Text: "!", Tag: "."},
				{Text: "!", Tag: "."},
				{Text: "All", Tag: "DT"},
				{Text: "special", Tag: "JJ"},
				{Text: "\x1b[1;31mCOLORED\x1b[0m", Tag: "CD"},
				{Text: "chars", Tag: "NNS"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagSpeech(tt.text)
			if !reflect.DeepEqual(result, tt.expected) && (len(result)+len(tt.expected)) != 0 {
				t.Errorf("\nResult: %v\nExpected: %v", result, tt.expected)
			}
		})
	}
}
