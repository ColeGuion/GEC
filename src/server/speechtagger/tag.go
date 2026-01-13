package speechtagger

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	none = regexp.MustCompile(`^(?:0|\*[\w?]\*|\*\-\d{1,3}|\*[A-Z]+\*\-\d{1,3}|\*)$`)
	keep = regexp.MustCompile(`^\-[A-Z]{3}\-$`)	// Three uppercase letters enclosed in `-` and `-`
)

func (pt *perceptronTagger) tag(tokens []*Token) []*Token {
	var tag string
	var found bool
	p1, p2 := "-START-", "-START2-"
	length := len(tokens) + 4

	// Create Context
	context := make([]string, length)
	context[0] = p1
	context[1] = p2
	for i, t := range tokens {
		context[i+2] = normalize(t.Text)
	}
	context[length-2] = "-END-"
	context[length-1] = "-END2-"

	// Tagging
	for i := 0; i < len(tokens); i++ {
		word := tokens[i].Text
		if word == "-" {
			tag = "-"
		} else if _, ok := emoticons[word]; ok {
			tag = "SYM"
		} else if strings.HasPrefix(word, "@") {
			tag = "NN"
		} else if none.MatchString(word) {
			tag = "-NONE-"
		} else if keep.MatchString(word) {
			tag = word
		} else if tag, found = pt.model.tagMap[word]; !found {
			tag = pt.model.predict(featurize(i, context, word, p1, p2))
		}
		tokens[i].Tag = tag
		p2 = p1
		p1 = tag
	}

	return tokens
}

// Predict the most likely tag for a given word
func (m *averagedPerceptron) predict(features map[string]float64) string {
	var weights map[string]float64
	var found bool

	scores := make(map[string]float64)
	for feat, value := range features {
		if weights, found = m.weights[feat]; !found || value == 0 {
			continue
		}
		for label, weight := range weights {
			scores[label] += value * weight
		}
	}
	return max(scores)
}

func featurize(i int, ctx []string, w, p1, p2 string) map[string]float64 {
	feats := make(map[string]float64)
	suf := min(len(w), 3)
	i = min(len(ctx)-2, i+2)
	iminus := min(len(ctx[i-1]), 3)
	iplus := min(len(ctx[i+1]), 3)
	feats = add([]string{"bias"}, feats)
	feats = add([]string{"i suffix", w[len(w)-suf:]}, feats)
	feats = add([]string{"i pref1", string(w[0])}, feats)
	feats = add([]string{"i-1 tag", p1}, feats)
	feats = add([]string{"i-2 tag", p2}, feats)
	feats = add([]string{"i tag+i-2 tag", p1, p2}, feats)
	feats = add([]string{"i word", ctx[i]}, feats)
	feats = add([]string{"i-1 tag+i word", p1, ctx[i]}, feats)
	feats = add([]string{"i-1 word", ctx[i-1]}, feats)
	feats = add([]string{"i-1 suffix", ctx[i-1][len(ctx[i-1])-iminus:]}, feats)
	feats = add([]string{"i-2 word", ctx[i-2]}, feats)
	feats = add([]string{"i+1 word", ctx[i+1]}, feats)
	feats = add([]string{"i+1 suffix", ctx[i+1][len(ctx[i+1])-iplus:]}, feats)
	feats = add([]string{"i+2 word", ctx[i+2]}, feats)
	return feats
}

func add(args []string, features map[string]float64) map[string]float64 {
	key := strings.Join(args, " ")
	features[key]++
	return features
}

// Find the key with the max score
func max(scores map[string]float64) string {
	var class string
	max := math.Inf(-1)
	for label, value := range scores {
		if value > max {
			max = value
			class = label
		}
	}
	return class
}

// Normalize a string
func normalize(word string) string {
	if word == "" {
		return word
	}
	first := string(word[0])
	if strings.Contains(word, "-") && first != "-" {
		return "!HYPHEN"
	} else if _, err := strconv.Atoi(word); err == nil && len(word) == 4 {
		return "!YEAR"
	} else if _, err := strconv.Atoi(first); err == nil {
		return "!DIGITS"
	}
	return strings.ToLower(word)
}
