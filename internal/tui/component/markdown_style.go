package component

import (
	"fmt"
	"image/color"

	"charm.land/glamour/v2/ansi"
)

// markdownStyle is the complete Glamour style of decision 29. Each colour is
// a palette token. Text without a colour is fg. Headings are accent and bold
// with no "#" prefix. Strong text is bold. Inline code and code blocks are
// code, with no background and no syntax colours; a code block has no border
// and no title and is indented by 2 cells. List items use "• " in accent.
// Links show their text in info; markdownLinks removes the URL. Tables have
// inner rules only, in border. The document margin is 0. Emphasis has no
// style, because the palette has no italic.
var markdownStyle = ansi.StyleConfig{
	Document: ansi.StyleBlock{Margin: uintPtr(0)},
	Heading: ansi.StyleBlock{
		StylePrimitive: ansi.StylePrimitive{Color: hexPtr(Accent), Bold: boolPtr(true), BlockSuffix: "\n"},
	},
	Paragraph: ansi.StyleBlock{},
	List:      ansi.StyleList{LevelIndent: 2},
	Strong:    ansi.StylePrimitive{Bold: boolPtr(true)},
	Item:      ansi.StylePrimitive{BlockPrefix: markdownBullet},
	Enumeration: ansi.StylePrimitive{
		BlockPrefix: ". ",
	},
	Task:     ansi.StyleTask{Ticked: "✓ ", Unticked: "○ "},
	LinkText: ansi.StylePrimitive{Color: hexPtr(Info)},
	Code:     ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: hexPtr(Code)}},
	CodeBlock: ansi.StyleCodeBlock{
		StyleBlock: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Color: hexPtr(Code)}, Margin: uintPtr(2)},
	},
	Table: ansi.StyleTable{
		CenterSeparator: strPtr(borderStyle.Render("┼")),
		ColumnSeparator: strPtr(borderStyle.Render("│")),
		RowSeparator:    strPtr(borderStyle.Render("─")),
	},
	HorizontalRule: ansi.StylePrimitive{Color: hexPtr(Border), Format: "────"},
}

// markdownBullet is the list item prefix in fg, as the golden-frames boards
// draw it.
const markdownBullet = "• "

func hexPtr(c color.Color) *string {
	r, g, b, _ := c.RGBA()
	return strPtr(fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8))
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func uintPtr(u uint) *uint    { return &u }
