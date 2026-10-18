package memory

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var errNoFrontmatter = errors.New("frontmatter: no block")

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
			return nil, fmt.Errorf("frontmatter: missing %s", key)
		case counts[key] > 1:
			return nil, fmt.Errorf("frontmatter: duplicate %s", key)
		case values[key] == "":
			return nil, fmt.Errorf("frontmatter: empty %s", key)
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
		return nil, fmt.Errorf("frontmatter: name %s differs from %s", displayValue(values["name"]), displayValue(name))
	}
	return values, nil
}

func splitFrontmatterLine(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ": ")
	if !ok {
		key, value, ok = strings.Cut(line, ":")
	}
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
