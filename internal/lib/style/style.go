// Package style holds the SGR palette that the graph and the picker share.
// Each state also has a glyph or a word, so colour only adds.
package style

// The SGR parameters of the palette.
const (
	Border   = "38;2;61;58;51"
	Dim      = "38;2;163;157;144"
	Accent   = "38;2;245;185;66"
	Selected = "1;38;2;245;185;66"
	Working  = "38;2;110;230;154"
	Error    = "38;2;232;112;95"
)

// The SGR parameters of the harness colours. Pi keeps the default foreground.
const (
	Claude = "38;2;217;119;87"
	Codex  = "38;2;166;227;201"
)

// Harness gives the colour of a harness name, or "" for the default
// foreground.
func Harness(name string) string {
	switch name {
	case "claude":
		return Claude
	case "codex":
		return Codex
	}
	return ""
}

// Color wraps text in the SGR style. An empty style keeps the text plain.
func Color(style, text string) string {
	if style == "" {
		return text
	}
	return "\x1b[" + style + "m" + text + "\x1b[0m"
}
