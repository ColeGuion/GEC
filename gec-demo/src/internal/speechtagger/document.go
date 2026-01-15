// src/internal/speechtagger/document.go
package speechtagger

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	//"runtime"
	"strings"

	"gec-demo/src/internal/print"

	"gopkg.in/neurosnap/sentences.v1"
	"gopkg.in/neurosnap/sentences.v1/data"
)

var (
	TaggerModel   *Model
	SentTokenizer *sentences.DefaultSentenceTokenizer
	TagsGob       string
	WeightsGob    string
)

// Resolve paths relative to repo root
func resolvePath(parts ...string) (string, error) {
	root := os.Getenv("GEC_ROOT")
	if root == "" {
		return "", fmt.Errorf("GEC_ROOT environment variable is not set")
	}
	return filepath.Join(append([]string{root}, parts...)...), nil
}

// Initialize the part-of-speech tagging model
func InitTaggingModel() error {
	var wts map[string]map[string]float64
	var tags map[string]string
	var err error

	// Resolve paths to local GEC_ROOT env variable
	TagsGob, err = resolvePath("src", "internal", "speechtagger", "data", "tags.gob")
	if err != nil {
		return err
	}
	WeightsGob, err = resolvePath("src", "internal", "speechtagger", "data", "weights.gob")
	if err != nil {
		return err
	}
	print.Info("Tags-Gob Path: \x1b[36m%q\x1b[0m", TagsGob)
	print.Info("Weights-Gob Path: \x1b[36m%q\x1b[0m", WeightsGob)

	// Decode the gob files
	err = decodeGob(TagsGob, &tags)
	if err != nil {
		return err
	}
	err = decodeGob(WeightsGob, &wts)
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
	b, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading file: %w", err)
	}

	dec := gob.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(obj); err != nil {
		return fmt.Errorf("Error decoding gob file: %w", err)
	}
	return nil
}
