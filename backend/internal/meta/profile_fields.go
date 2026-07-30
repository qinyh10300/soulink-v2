package meta

import (
	"strings"
	"unicode/utf8"
)

const MaxMajorDescriptionLength = 20

func IsValidMajorDescription(value string) bool {
	trimmed := strings.TrimSpace(value)
	length := utf8.RuneCountInString(trimmed)
	return length >= 1 && length <= MaxMajorDescriptionLength
}
