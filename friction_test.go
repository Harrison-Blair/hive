package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const frictionRoot = "reference/dogfood/friction"

// TestFrictionLogMatchesIndex checks that FRICTION.md and the entry files of
// the friction log agree.
func TestFrictionLogMatchesIndex(t *testing.T) {
	for _, problem := range checkFriction(frictionRoot) {
		t.Error(problem)
	}
}

const validEntry = `---
title: "A spawn fails"
description: "the first tab is busy"
status: "Open"
opened: 2026-10-29
decision: ""
---

**Issue:** text.
`

const validResolved = `---
title: "A stop fails"
description: "the pane stays open"
status: "Resolved 2026-10-29"
opened: 2026-10-28
decision: "Fixed in PR 60"
---

**Issue:** text.
`

const validIndex = `# Friction log

Rules text.

## Open

- [A spawn fails](entries/spawn-fails.md) — the first tab is busy

## Resolved

- [A stop fails](entries/resolved/stop-fails.md) — the pane stays open
`

// TestCheckFrictionFindsBrokenLogs checks that the checker reports each kind
// of disagreement, and reports nothing for a valid log.
func TestCheckFrictionFindsBrokenLogs(t *testing.T) {
	cases := []struct {
		name  string
		index string
		files map[string]string
		want  string
	}{
		{name: "valid", index: validIndex, want: ""},
		{
			name:  "line without file",
			index: validIndex + "- [Gone](entries/resolved/gone.md) — no file\n",
			want:  "entries/resolved/gone.md: index line has no entry file",
		},
		{
			name:  "file without line",
			index: validIndex,
			files: map[string]string{"entries/extra.md": strings.Replace(validEntry, "A spawn fails", "Extra", 1)},
			want:  "entries/extra.md: entry file has no index line",
		},
		{
			name:  "duplicate line",
			index: validIndex + "- [A stop fails](entries/resolved/stop-fails.md) — the pane stays open\n",
			want:  "entries/resolved/stop-fails.md: duplicate index line",
		},
		{
			name:  "title mismatch",
			index: strings.Replace(validIndex, "[A spawn fails]", "[A spawn breaks]", 1),
			want:  `entries/spawn-fails.md: index title "A spawn breaks" differs from title "A spawn fails"`,
		},
		{
			name:  "description mismatch",
			index: strings.Replace(validIndex, "— the first tab is busy", "— busy", 1),
			want:  `entries/spawn-fails.md: index description "busy" differs from description "the first tab is busy"`,
		},
		{
			name:  "resolved line points into entries",
			index: strings.Replace(validIndex, "entries/resolved/stop-fails.md", "entries/stop-fails.md", 1),
			files: map[string]string{"entries/stop-fails.md": validResolved},
			want:  "entries/stop-fails.md: line under ## Resolved must point into entries/resolved/",
		},
		{
			name:  "open line points into resolved",
			index: strings.Replace(validIndex, "entries/spawn-fails.md", "entries/resolved/spawn-fails.md", 1),
			files: map[string]string{"entries/resolved/spawn-fails.md": validEntry},
			want:  "entries/resolved/spawn-fails.md: line under ## Open must point into entries/ outside entries/resolved/",
		},
		{
			name:  "open status under resolved",
			index: validIndex,
			files: map[string]string{"entries/resolved/stop-fails.md": strings.Replace(validResolved, `"Resolved 2026-10-29"`, `"Open"`, 1)},
			want:  `entries/resolved/stop-fails.md: status "Open" under ## Resolved must be Resolved YYYY-MM-DD`,
		},
		{
			name:  "resolved status under open",
			index: validIndex,
			files: map[string]string{"entries/spawn-fails.md": strings.Replace(validEntry, `"Open"`, `"Resolved 2026-10-29"`, 1)},
			want:  `entries/spawn-fails.md: status "Resolved 2026-10-29" under ## Open must be Open`,
		},
		{
			name:  "missing field",
			index: validIndex,
			files: map[string]string{"entries/spawn-fails.md": strings.Replace(validEntry, "decision: \"\"\n", "", 1)},
			want:  "entries/spawn-fails.md: frontmatter has no decision field",
		},
		{
			name:  "bad opened date",
			index: validIndex,
			files: map[string]string{"entries/spawn-fails.md": strings.Replace(validEntry, "2026-10-29", "yesterday", 1)},
			want:  `entries/spawn-fails.md: opened "yesterday" must be YYYY-MM-DD`,
		},
		{
			name:  "no frontmatter",
			index: validIndex,
			files: map[string]string{"entries/spawn-fails.md": "**Issue:** text.\n"},
			want:  "entries/spawn-fails.md: missing YAML frontmatter",
		},
		{
			name:  "malformed line",
			index: strings.Replace(validIndex, " — the first tab is busy", "", 1),
			want:  "FRICTION.md:7: entry line must be - [TITLE](entries/SLUG.md) — DESCRIPTION",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			files := map[string]string{
				"FRICTION.md":                    tc.index,
				"entries/spawn-fails.md":         validEntry,
				"entries/resolved/stop-fails.md": validResolved,
			}
			for path, content := range tc.files {
				files[path] = content
			}
			for path, content := range files {
				full := filepath.Join(root, path)
				if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			problems := checkFriction(root)
			if tc.want == "" {
				if len(problems) != 0 {
					t.Fatalf("problems for a valid log: %q", problems)
				}
				return
			}
			for _, problem := range problems {
				if problem == tc.want {
					return
				}
			}
			t.Fatalf("problems %q do not contain %q", problems, tc.want)
		})
	}
}

var (
	frictionLinePattern = regexp.MustCompile(`^- \[([^\]]+)\]\((entries/[^)]+\.md)\) — (.+)$`)
	frictionDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

type frictionLine struct {
	title       string
	description string
	section     string
}

// checkFriction returns each disagreement between FRICTION.md in root and the
// entry files below root/entries/.
func checkFriction(root string) []string {
	index, err := os.ReadFile(filepath.Join(root, "FRICTION.md"))
	if err != nil {
		return []string{"read friction index: " + err.Error()}
	}
	var problems []string
	lines := map[string]frictionLine{}
	section := ""
	for number, line := range strings.Split(string(index), "\n") {
		if heading, ok := strings.CutPrefix(line, "## "); ok {
			section = strings.TrimSpace(heading)
			continue
		}
		if section != "Open" && section != "Resolved" || !strings.HasPrefix(line, "- ") {
			continue
		}
		match := frictionLinePattern.FindStringSubmatch(line)
		if match == nil {
			problems = append(problems, fmt.Sprintf("FRICTION.md:%d: entry line must be - [TITLE](entries/SLUG.md) — DESCRIPTION", number+1))
			continue
		}
		path := match[2]
		if _, exists := lines[path]; exists {
			problems = append(problems, path+": duplicate index line")
			continue
		}
		lines[path] = frictionLine{title: match[1], description: match[3], section: section}
	}

	entries := map[string]map[string]string{}
	err = filepath.WalkDir(filepath.Join(root, "entries"), func(path string, item fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if item.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fields, err := frictionFrontmatter(string(content))
		if err != nil {
			problems = append(problems, relative+": "+err.Error())
			fields = nil
		}
		entries[relative] = fields
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		problems = append(problems, "read friction entries: "+err.Error())
	}

	for path, line := range lines {
		fields, ok := entries[path]
		if !ok {
			problems = append(problems, path+": index line has no entry file")
			continue
		}
		inResolved := strings.HasPrefix(path, "entries/resolved/")
		if line.section == "Resolved" && !inResolved {
			problems = append(problems, path+": line under ## Resolved must point into entries/resolved/")
		}
		if line.section == "Open" && inResolved {
			problems = append(problems, path+": line under ## Open must point into entries/ outside entries/resolved/")
		}
		if fields == nil {
			continue
		}
		if line.title != fields["title"] {
			problems = append(problems, fmt.Sprintf("%s: index title %q differs from title %q", path, line.title, fields["title"]))
		}
		if line.description != fields["description"] {
			problems = append(problems, fmt.Sprintf("%s: index description %q differs from description %q", path, line.description, fields["description"]))
		}
		status := fields["status"]
		date, resolved := strings.CutPrefix(status, "Resolved ")
		if line.section == "Open" && status != "Open" {
			problems = append(problems, fmt.Sprintf("%s: status %q under ## Open must be Open", path, status))
		}
		if line.section == "Resolved" && (!resolved || !frictionDatePattern.MatchString(date)) {
			problems = append(problems, fmt.Sprintf("%s: status %q under ## Resolved must be Resolved YYYY-MM-DD", path, status))
		}
	}
	for path := range entries {
		if _, ok := lines[path]; !ok {
			problems = append(problems, path+": entry file has no index line")
		}
	}
	return problems
}

// frictionFrontmatter returns the fields of the YAML frontmatter of one entry,
// and checks that each required field is present.
func frictionFrontmatter(content string) (map[string]string, error) {
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		return nil, errors.New("missing YAML frontmatter")
	}
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return nil, errors.New("unterminated YAML frontmatter")
	}
	fields := map[string]string{}
	for _, line := range strings.Split(rest[:end], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
	}
	for _, key := range []string{"title", "description", "status", "opened", "decision"} {
		if _, ok := fields[key]; !ok {
			return nil, fmt.Errorf("frontmatter has no %s field", key)
		}
	}
	if !frictionDatePattern.MatchString(fields["opened"]) {
		return nil, fmt.Errorf("opened %q must be YYYY-MM-DD", fields["opened"])
	}
	return fields, nil
}
