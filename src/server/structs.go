// structs.go
package main

// ********* SERVER ENDPOINT *********
type GecRequest struct {
	Text string `json:"text"`
}

/*
	 type GecResponse struct {
		CorrectedText string   `json:"corrected_text"`
		TextMarkup    []Markup `json:"textMarkup"`
	}
*/
type GecResponse struct {
	CorrectedText       string        `json:"corrected_text"`
	TextMarkups         []Markup      `json:"text_markups"`
	GibberishScores     []GibbResults `json:"gibberish_scores"`
	CharacterCount      int           `json:"character_count"`
	ErrorCharacterCount int           `json:"error_character_count"`
	ContainsProfanity   bool          `json:"contains_profanity"`
	ServiceTime         float64       `json:"service_time"`
}

// ********* GEC *********
type Markup struct {
	Index   int    `json:"index"`
	Length  int    `json:"length"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type Misspell struct {
	Index       int      // The index of the misspelled word in the text
	Length      int      // The length of the misspelled word
	Type        string   // The type of error
	Suggestions []string // Suggested word replacements
}

type GrammarResult struct {
	CorrectText string
	GibbScores  []GibbResults
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

type GibbResults struct {
	Index  int        `json:"index"`
	Length int        `json:"length"`
	Score  GibbScores `json:"score"`
}
type GibbScores struct {
	Clean     float32 `json:"clean"`
	Mild      float32 `json:"mild"`
	Noise     float32 `json:"noise"`
	WordSalad float32 `json:"wordSalad"`
}

/* type Wrapper struct {
	TextMarkups     []Markup `json:"textMarkups"`
	ErrorCharacters int      `json:"errorCharacters"`
	ProfaneWords    []string `json:"profaneWords"`
}
 */