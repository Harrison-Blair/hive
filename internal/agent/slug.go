package agent

import "strings"

// Slug turns text into a part of an agent name of at most max bytes. It
// lowercases text, changes each run of characters outside [a-z0-9_-] to one
// hyphen, and trims the hyphens at both ends, also after it cuts the text to
// max. The result can be empty.
func Slug(text string, max int) string {
	var out strings.Builder
	run := false
	for _, r := range strings.ToLower(text) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			out.WriteRune(r)
			run = false
			continue
		}
		if !run {
			out.WriteByte('-')
			run = true
		}
	}
	slug := strings.Trim(out.String(), "-")
	if len(slug) > max {
		slug = strings.TrimRight(slug[:max], "-")
	}
	return slug
}
