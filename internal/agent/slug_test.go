package agent

import (
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := []struct {
		text string
		max  int
		want string
	}{
		{"hive", 64, "hive"},
		{"My Repo", 64, "my-repo"},
		{"ror.desktop", 64, "ror-desktop"},
		{"a  ..b", 64, "a-b"},
		{"keep_under-score", 64, "keep_under-score"},
		{"--edge--", 64, "edge"},
		{"Ünï", 64, "n"},
		{"日本", 64, ""},
		{"/", 64, ""},
		{"abcdef", 3, "abc"},
		{"ab cdef", 3, "ab"},
	}
	for _, test := range cases {
		if got := Slug(test.text, test.max); got != test.want {
			t.Errorf("Slug(%q, %d) = %q; want %q", test.text, test.max, got, test.want)
		}
	}
	if got := Slug(strings.Repeat("x", 100), 64); got != strings.Repeat("x", 64) {
		t.Errorf("long slug = %q", got)
	}
}
