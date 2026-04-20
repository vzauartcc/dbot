package helpers

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

var notCapitalized = []string{"a", "and", "as", "at", "but", "by", "down", "for", "from", "if", "in", "into", "like", "near", "nor", "of", "off", "on", "once", "onto", "or", "over", "past", "so", "than", "that", "the", "to", "upon", "when", "with", "yet"}

func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	r, size := utf8.DecodeRuneInString(s)

	return string(unicode.ToUpper(r)) + s[size:]
}

func TitleCase(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if i == 0 {
			words[i] = CapitalizeFirst(word)
			continue
		}

		if slices.Contains(notCapitalized, strings.ToLower(words[i])) {
			words[i] = strings.ToLower(words[i])
			continue
		}

		words[i] = CapitalizeFirst(word)
	}

	return strings.Join(words, " ")
}
