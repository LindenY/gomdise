package mdl

import (
	"strings"
	"unicode"
)

type tagOptions string

func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1])
	}
	return tag, tagOptions("")
}

func isValidTag(tag string) bool {
	if tag == "" {
		return false
	}
	for _, c := range tag {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
		// Backslash and quote chars are reserved, but
		// otherwise any punctuation chars are allowed
		// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}
