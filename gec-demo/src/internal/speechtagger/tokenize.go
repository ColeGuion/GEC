package speechtagger

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	specialRgx = regexp.MustCompile(`^(?:[A-Za-z]\.){2,}$|^[A-Z][a-z]{1,2}\.$`)
	sanitizer  = strings.NewReplacer(
		"\u201c", `"`,
		"\u201d", `"`,
		"\u2018", "'",
		"\u2019", "'",
		"&rsquo;", "'",
	)
	contractions = []string{"'ll", "'s", "'re", "'m", "n't"}
	suffixes     = []string{",", ")", `"`, "]", "!", ";", ".", "?", ":", "'"}
	prefixes     = []string{"$", "(", `"`, "["}
	emoticons    = map[string]int{
		"(-8":         1,
		"(-;":         1,
		"(-_-)":       1,
		"(._.)":       1,
		"(:":          1,
		"(=":          1,
		"(o:":         1,
		"(¬_¬)":       1,
		"(ಠ_ಠ)":       1,
		"(╯°□°）╯︵┻━┻": 1,
		"-__-":        1,
		"8-)":         1,
		"8-D":         1,
		"8D":          1,
		":(":          1,
		":((":         1,
		":(((":        1,
		":()":         1,
		":)))":        1,
		":-)":         1,
		":-))":        1,
		":-)))":       1,
		":-*":         1,
		":-/":         1,
		":-X":         1,
		":-]":         1,
		":-o":         1,
		":-p":         1,
		":-x":         1,
		":-|":         1,
		":-}":         1,
		":0":          1,
		":3":          1,
		":P":          1,
		":]":          1,
		":`(":         1,
		":`)":         1,
		":`-(":        1,
		":o":          1,
		":o)":         1,
		"=(":          1,
		"=)":          1,
		"=D":          1,
		"=|":          1,
		"@_@":         1,
		"O.o":         1,
		"O_o":         1,
		"V_V":         1,
		"XDD":         1,
		"[-:":         1,
		"^___^":       1,
		"o_0":         1,
		"o_O":         1,
		"o_o":         1,
		"v_v":         1,
		"xD":          1,
		"xDD":         1,
		"¯\\(ツ)/¯":    1,
	}
)

// Splits a sentence into a slice of Tokens with only the Text field filled
func Tokenize(text string) []*Token {
	var tokens []*Token
	cache := map[string][]*Token{}   // Cache for storing tokenized strings, so that we don't tokenize the same string multiple times
	clean := sanitizer.Replace(text) // Replace string pairings with the sanitizer
	length := len(clean)
	white := false

	start, index := 0, 0
	for index <= length {
		uc, size := utf8.DecodeRuneInString(clean[index:])
		if size == 0 {
			break // End of string
		} else if index == 0 {
			white = unicode.IsSpace(uc)
		}

		if unicode.IsSpace(uc) != white {
			if start < index {
				span := clean[start:index]
				if toks, found := cache[span]; found {
					tokens = append(tokens, toks...)
				} else {
					toks := doSplit(span)
					cache[span] = toks
					tokens = append(tokens, toks...)
				}
			}
			start = index
			if uc == ' ' {
				start++ // Increment
			}
			white = !white
		}
		index += size
	}

	if start < index {
		tokens = append(tokens, doSplit(clean[start:index])...)
	}

	return tokens
}

// Splits the given token into a slice of Token pointers based on various conditions
func doSplit(token string) []*Token {
	tokens := []*Token{}
	suffs := []*Token{}

	last := 0
	for token != "" && utf8.RuneCountInString(token) != last {
		_, found := emoticons[token]
		if found || specialRgx.MatchString(token) {
			// Special Case	 (Emoticons or special regex match)
			// Add the token without further processing
			tokens = addToken(token, tokens)
			break
		}

		last = utf8.RuneCountInString(token)
		lower := strings.ToLower(token)
		if hasAnyPrefix(token, prefixes) {
			// Remove Prefixes (e.g. $100 -> [$, 100])
			tokens = addToken(string(token[0]), tokens)
			token = token[1:]
		} else if idx := hasAnyIndex(lower, contractions); idx > -1 {
			// Contractions (e.g. "they'll" -> ["they", "'ll"])
			tokens = addToken(token[:idx], tokens)
			token = token[idx:]
		} else if hasAnySuffix(token, suffixes) {
			// Remove Suffixes (e.g. "Well)" -> ["Well", ","])
			suffs = append([]*Token{{Text: string(token[len(token)-1])}}, suffs...)
			token = token[:len(token)-1]
		} else {
			tokens = addToken(token, tokens)
		}
	}

	return append(tokens, suffs...)
}

// Appends a non-empty, trimmed token to a slice of tokens
func addToken(s string, toks []*Token) []*Token {
	if strings.TrimSpace(s) != "" {
		toks = append(toks, &Token{Text: s})
	}
	return toks
}

// Checks if a string starts with any of the provided prefixes
func hasAnyPrefix(s string, prefixes []string) bool {
	n := len(s)
	for _, prefix := range prefixes {
		if n > len(prefix) && strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// Checks if a string ends with any of the provided suffixes
func hasAnySuffix(s string, suffixes []string) bool {
	n := len(s)
	for _, suffix := range suffixes {
		if n > len(suffix) && strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}

// Checks if any of the provided suffixes are present in the given string and returns the index of the first occurrence of any suffix found
// Ensures the length of `s` is greater than the length of the suffix before returning the index
func hasAnyIndex(s string, suffixes []string) int {
	n := len(s)
	for _, suffix := range suffixes {
		idx := strings.Index(s, suffix)
		if idx >= 0 && n > len(suffix) {
			return idx
		}
	}
	return -1
}
