package component

import (
	"bytes"
	"regexp"
	"strings"

	glamouransi "charm.land/glamour/v2/ansi"
	"github.com/charmbracelet/x/ansi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// RenderMarkdown renders an agent reply with Glamour and the style of
// decision 29, wrapped at width cells (decision 28). It removes only the
// padding that Glamour adds: the blank first line and the trailing spaces of
// each line (decision 31). It keeps leading spaces. Glamour wraps a list
// with breakpoints, and a final "." or "," can then hang one cell past the
// width; such a line is wrapped again at width.
func RenderMarkdown(src string, width int) ([]string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.Strikethrough, extension.TaskList),
		goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(markdownLinks{}, 100))),
	)
	// SetRenderer after New drops the HTML renderers of the extensions, as
	// glamour.NewTermRenderer does.
	md.SetRenderer(renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(
		glamouransi.NewRenderer(glamouransi.Options{WordWrap: width, PreserveNewLines: true, Styles: markdownStyle}), 1000))))
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) > 0 && strings.TrimSpace(ansi.Strip(lines[0])) == "" {
		lines = lines[1:]
	}
	var out []string
	for _, l := range lines {
		l = trimRightSpaces(l)
		if ansi.StringWidth(l) > width {
			out = append(out, strings.Split(ansi.Wordwrap(l, width, ""), "\n")...)
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

// sgr matches one SGR escape sequence, and trailingSpaces matches the run
// of spaces and SGR sequences at the end of a line.
var (
	sgr            = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	trailingSpaces = regexp.MustCompile(`(?:\x1b\[[0-9;]*m| )+$`)
)

// trimRightSpaces removes the trailing spaces of a line and keeps each
// escape sequence among them, so a style that ends after the spaces still
// ends.
func trimRightSpaces(line string) string {
	loc := trailingSpaces.FindStringIndex(line)
	if loc == nil {
		return line
	}
	return line[:loc[0]] + strings.Join(sgr.FindAllString(line[loc[0]:], -1), "")
}

// markdownLinks sets the destination of each link to "#", and puts a link
// with the destination "#" in place of each autolink. Glamour then draws the
// link text in the LinkText style with no URL and no OSC 8 hyperlink
// (decision 29). The text of an autolink is its URL or its address.
type markdownLinks struct{}

func (markdownLinks) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	var autos []*ast.AutoLink
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch l := n.(type) {
		case *ast.Link:
			if entering {
				l.Destination = []byte("#")
			}
		case *ast.AutoLink:
			if entering {
				autos = append(autos, l)
			}
		}
		return ast.WalkContinue, nil
	})
	for _, a := range autos {
		// The label is a slice of src, so its offset is the difference of
		// the capacities.
		label := a.Label(src)
		start := cap(src) - cap(label)
		link := ast.NewLink()
		link.Destination = []byte("#")
		link.AppendChild(link, ast.NewTextSegment(text.NewSegment(start, start+len(label))))
		a.Parent().ReplaceChild(a.Parent(), a, link)
	}
}
