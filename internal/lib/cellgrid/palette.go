package cellgrid

// Token is a palette token of decision 26.
type Token uint8

// The nine palette tokens. FG is the zero value: the terminal default.
const (
	FG Token = iota
	Border
	Dim
	Accent
	Working
	Done
	Info
	User
	Code
)

// palette is the one table of the tokens: name, grid file letter, and hex.
// The fg hex is the value that a board render uses. The converter matches
// only the default foreground to fg, never this hex.
var palette = [...]struct {
	name   string
	letter byte
	hex    string
}{
	FG:      {"fg", 'f', "#e6e1d6"},
	Border:  {"border", 'r', "#3d3a33"},
	Dim:     {"dim", 'd', "#a39d90"},
	Accent:  {"accent", 'a', "#f5b942"},
	Working: {"working", 'w', "#6ee69a"},
	Done:    {"done", 'o', "#7ccf6a"},
	Info:    {"info", 'i', "#6fb7ff"},
	User:    {"user", 'u', "#e9cf9b"},
	Code:    {"code", 'c', "#9fd0ff"},
}

// String returns the token name, for example "accent".
func (t Token) String() string { return palette[t].name }

// Hex returns the token hex, for example "#f5b942".
func (t Token) Hex() string { return palette[t].hex }
