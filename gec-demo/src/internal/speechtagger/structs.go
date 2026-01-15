package speechtagger

type Document struct {
	Model  *Model
	Text   string
	Tokens []*Token
}
type Token struct {
	Tag  string // The token's part-of-speech tag
	Text string // The token's actual content
}
type Model struct {
	Name   string
	tagger *perceptronTagger
}

// Port of Textblob's "fast and accurate" POS tagger
type perceptronTagger struct {
	model *averagedPerceptron
}

// Averaged Perceptron classifier
type averagedPerceptron struct {
	tagMap  map[string]string
	weights map[string]map[string]float64
}
