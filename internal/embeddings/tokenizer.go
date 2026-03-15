package embeddings

import (
	"strings"
	"unicode"
)

const (
	defaultMaxTokens = 128
)

// tokenize performs simplified tokenization for the all-MiniLM-L6-v2 model.
// It lowercases, splits on whitespace and punctuation, and truncates to maxTokens.
func tokenize(text string, maxTokens int) []string {
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	text = strings.ToLower(text)

	var tokens []string
	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	for _, r := range text {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			current.WriteRune(r)
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			flush()
			tokens = append(tokens, string(r))
		default:
			// whitespace and other characters act as separators
			flush()
		}
	}
	flush()

	if len(tokens) > maxTokens {
		tokens = tokens[:maxTokens]
	}

	return tokens
}

// padTokens pads or truncates a token sequence to the given length.
// Returns the padded tokens and an attention mask (1 for real, 0 for padding).
func padTokens(tokens []string, length int) (padded []string, attentionMask []int32) {
	padded = make([]string, length)
	attentionMask = make([]int32, length)

	for i := 0; i < length; i++ {
		if i < len(tokens) {
			padded[i] = tokens[i]
			attentionMask[i] = 1
		} else {
			padded[i] = "[PAD]"
			attentionMask[i] = 0
		}
	}

	return padded, attentionMask
}
