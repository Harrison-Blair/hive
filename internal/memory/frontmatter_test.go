package memory

import (
	"strings"
	"testing"
)

func TestParseFrontmatterValid(t *testing.T) {
	data := []byte("---\nname: base-is-dev\ndescription: The base branch of hive is dev.\nextra: ignored\n---\nBody.\n")
	values, err := parseFrontmatter(data)
	if err != nil {
		t.Fatalf("parseFrontmatter() error = %v; want success", err)
	}
	if values["name"] != "base-is-dev" {
		t.Fatalf("name = %q; want base-is-dev", values["name"])
	}
	if values["description"] != "The base branch of hive is dev." {
		t.Fatalf("description = %q", values["description"])
	}
}

func TestParseFrontmatterKeepsColonInValue(t *testing.T) {
	values, err := parseFrontmatter([]byte("---\nname: fact\ndescription: A: B\n---\n"))
	if err != nil {
		t.Fatalf("parseFrontmatter() error = %v; want success", err)
	}
	if values["description"] != "A: B" {
		t.Fatalf("description = %q; want A: B", values["description"])
	}
}

func TestParseFrontmatterRequiresColonSpace(t *testing.T) {
	cases := []struct {
		name string
		data string
		want string
	}{
		{"missing name", "---\nname:fact\ndescription: D.\n---\n", "frontmatter: missing name"},
		{"missing description", "---\nname: fact\ndescription:D.\n---\n", "frontmatter: missing description"},
		{"both", "---\nname:fact\ndescription:D.\n---\n", "frontmatter: missing name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseFrontmatter([]byte(tc.data))
			if err == nil || err.Error() != tc.want {
				t.Fatalf("parseFrontmatter() error = %v; want %q", err, tc.want)
			}
		})
	}
}

func TestParseFrontmatterTrimsAfterColonSpace(t *testing.T) {
	values, err := parseFrontmatter([]byte("---\nname:  fact  \ndescription: \t A fact. \t\nextra: ignored\n---\n"))
	if err != nil {
		t.Fatalf("parseFrontmatter() error = %v; want success", err)
	}
	if values["name"] != "fact" || values["description"] != "A fact." {
		t.Fatalf("values = %#v; want trimmed values", values)
	}
}

func TestParseFrontmatterErrors(t *testing.T) {
	cases := []struct {
		name string
		data string
		want string
	}{
		{"no block", "No frontmatter.\n", "frontmatter: no block"},
		{"no close", "---\nname: fact\ndescription: A fact.\n", "frontmatter: no block"},
		{"missing name", "---\ndescription: A fact.\n---\n", "frontmatter: missing name"},
		{"missing description", "---\nname: fact\n---\n", "frontmatter: missing description"},
		{"empty name", "---\nname:   \ndescription: A fact.\n---\n", "frontmatter: empty name"},
		{"empty description", "---\nname: fact\ndescription:   \n---\n", "frontmatter: empty description"},
		{"duplicate name", "---\nname: fact\nname: other\ndescription: A fact.\n---\n", "frontmatter: duplicate name"},
		{"duplicate description", "---\nname: fact\ndescription: A fact.\ndescription: Other.\n---\n", "frontmatter: duplicate description"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseFrontmatter([]byte(tc.data))
			if err == nil || err.Error() != tc.want {
				t.Fatalf("parseFrontmatter() error = %v; want %q", err, tc.want)
			}
		})
	}
}

func TestParseMemoryNameDiffers(t *testing.T) {
	_, err := parseMemory([]byte("---\nname: other\ndescription: A fact.\n---\n"), "fact")
	if err == nil || err.Error() != "frontmatter: name other differs from fact" {
		t.Fatalf("parseMemory() error = %v; want name difference", err)
	}
}

func TestParseMemoryNameDifferenceStaysOnOneLine(t *testing.T) {
	_, err := parseMemory([]byte("---\nname: bad\nname2: x\ndescription: A fact.\n---\n"), "fact")
	if err == nil {
		t.Fatal("parseMemory() accepted a missing name")
	}
	_, err = parseMemory([]byte("---\nname: bad\nname: x\ndescription: A fact.\n---\n"), "fact")
	if err == nil || !strings.Contains(err.Error(), "duplicate name") {
		t.Fatalf("parseMemory() error = %v; want duplicate name", err)
	}
}
