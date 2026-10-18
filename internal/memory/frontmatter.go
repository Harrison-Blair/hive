package memory

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type frontmatterError struct {
	category string
	message  string
}

func (e *frontmatterError) Error() string { return e.message }

// LogCategory returns the safe log text for one frontmatter error.
func (e *frontmatterError) LogCategory() string { return e.category }

var errNoFrontmatter = &frontmatterError{category: "frontmatter: no block", message: "frontmatter: no block"}

// parseFrontmatter parses one memory frontmatter block.
func parseFrontmatter(data []byte) (map[string]string, error) {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil, errNoFrontmatter
	}
	end := -1
	for index := 1; index < len(lines); index++ {
		if lines[index] == "---" {
			end = index
			break
		}
	}
	if end == -1 {
		return nil, errNoFrontmatter
	}
	counts := map[string]int{}
	values := map[string]string{}
	for _, line := range lines[1:end] {
		key, value, ok := splitFrontmatterLine(line)
		if !ok || key == "" {
			continue
		}
		counts[key]++
		values[key] = value
	}
	for _, key := range []string{"name", "description"} {
		switch {
		case counts[key] == 0:
			return nil, newFrontmatterError("frontmatter: missing "+key, "frontmatter: missing "+key)
		case counts[key] > 1:
			return nil, newFrontmatterError("frontmatter: duplicate "+key, "frontmatter: duplicate "+key)
		case values[key] == "":
			return nil, newFrontmatterError("frontmatter: empty "+key, "frontmatter: empty "+key)
		}
	}
	return values, nil
}

// parseMemory parses frontmatter and checks the memory name.
func parseMemory(data []byte, name string) (map[string]string, error) {
	values, err := parseFrontmatter(data)
	if err != nil {
		return nil, err
	}
	if values["name"] != name {
		message := fmt.Sprintf("frontmatter: name %s differs from %s", displayValue(values["name"]), displayValue(name))
		return nil, newFrontmatterError("frontmatter: name differs", message)
	}
	return values, nil
}

func newFrontmatterError(category, message string) error {
	return &frontmatterError{category: category, message: message}
}

func splitFrontmatterLine(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ": ")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), strings.TrimSpace(value), true
}

func displayValue(value string) string {
	if !utf8.ValidString(value) {
		return fmt.Sprintf("%q", value)
	}
	for _, char := range value {
		if unicode.IsControl(char) {
			return fmt.Sprintf("%q", value)
		}
	}
	return value
}
