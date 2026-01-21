// src/internal/gec/structs.go
package gec

// ********* SERVER ENDPOINT *********
type GecRequest struct {
	Text string `json:"text"`
}

type GecResponse struct {
	CorrectedText       string        `json:"corrected_text"`
	TextMarkups         []Markup      `json:"text_markups"`
	CharacterCount      int           `json:"character_count"`
	ErrorCharacterCount int           `json:"error_character_count"`
	ContainsProfanity   bool          `json:"contains_profanity"`
	ServiceTime         float64       `json:"service_time"`
}

// ********* GEC *********
type Markup struct {
	Index    int    `json:"index"`
	Length   int    `json:"length"`
	Message  string `json:"message"`
	Category string `json:"category"`
}

type Misspell struct {
	Index       int      // The index of the misspelled word in the text
	Length      int      // The length of the misspelled word
	Category    string   // The type of error
	Suggestions []string // Suggested word replacements
}

type GrammarResult struct {
	CorrectText string
	GpuId       int
	Err         error
	ServiceTime float64
}

type WorkItem struct {
	Count    int
	Text     string
	AllTexts []string
	Ch       chan GrammarResult
}
