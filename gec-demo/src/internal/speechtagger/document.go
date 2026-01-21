// src/internal/speechtagger/document.go
package speechtagger

import (
	"bytes"
	"embed"
	"encoding/gob"
	"fmt"
	"strings"

	"gopkg.in/neurosnap/sentences.v1"
	"gopkg.in/neurosnap/sentences.v1/data"
)

//go:embed data/*.gob
var modelData embed.FS

var (
	TaggerModel   *Model
	SentTokenizer *sentences.DefaultSentenceTokenizer
)

// Initialize the part-of-speech tagging model
func InitTaggingModel() error {
	var wts map[string]map[string]float64
	var tags map[string]string
	var err error

	// Decode the gob files
	err = decodeGob("data/tags.gob", &tags)
	if err != nil {
		return err
	}
	err = decodeGob("data/weights.gob", &wts)
	if err != nil {
		return err
	}

	// Create a new AveragedPerceptron model
	percepMod := averagedPerceptron{
		tagMap:  tags,
		weights: wts,
	}

	TaggerModel = &Model{
		Name:   "en-v2.0.0",
		tagger: &perceptronTagger{model: &percepMod},
	}

	// Load the Sentence Tokenizer
	b, err := data.Asset("data/english.json")
	if err != nil {
		return fmt.Errorf("Error loading english data for sentence tokenizer: %w", err)
	}
	// Load the training data
	training, err := sentences.LoadTraining(b)
	if err != nil {
		return fmt.Errorf("Error loading training data for sentence tokenizer: %w", err)
	}
	// Create the sentence tokenizer
	SentTokenizer = sentences.NewSentenceTokenizer(training)
	return nil
}

// Tag parts-of-speech with tokens and tags
func TagSpeech(text string) []*Token {
	// Split text into tokens
	var tokens []*Token
	tokens = append(tokens, Tokenize(text)...)

	// Add the tags to the tokens
	tokens = TaggerModel.tagger.tag(tokens)
	return tokens
}

// Split text into sentences
func SplitBySentences(text string) (allTexts []string) {
	sentences := SentTokenizer.Tokenize(text)
	sents := make([]string, len(sentences))
	for i, s := range sentences {
		sents[i] = strings.TrimSpace(s.Text)
	}
	return sents
}

// Decode .gob file
func decodeGob(filePath string, obj any) error {
	// Use embedded path
	data, err := modelData.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed reading embedded gob file %q: %w", filePath, err)
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(obj); err != nil {
		return fmt.Errorf("Error decoding gob file: %w", err)
	}
	return nil
}
