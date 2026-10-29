package project

import (
	"fmt"
	"unicode"
)

const invalidIDRule = "an ID matches ^[a-z0-9]+(-[a-z0-9]+)*$ and has at most 18 characters"

// InvalidIDMessage returns the error text of an invalid project ID.
func InvalidIDMessage(id string) string {
	if HasControl(id) {
		id = fmt.Sprintf("%q", id)
	}
	return fmt.Sprintf("invalid project id: %s: %s", id, invalidIDRule)
}

// HasControl reports whether value contains a control character.
func HasControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
