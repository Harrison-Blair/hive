package tui

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Harrison-Blair/hive/internal/tui/component"
)

// logLabelCells is the label column of the field panel (logs.md rule 4).
// The value column starts after it.
const logLabelCells = 14

// logFirstFields are the fields that the field panel shows first, in this
// order. The other fields follow in the order of the event.
var logFirstFields = []string{"time", "level", "msg", "component"}

// errBadLogObject is a nested object that does not read.
var errBadLogObject = errors.New("bad nested object")

var errorTextStyle = lipgloss.NewStyle().Foreground(component.Error)

// logFieldNode is one field of an event: a key with a scalar value, or a key
// with the fields of a nested object.
type logFieldNode struct {
	key    string
	value  string
	nested []logFieldNode
	object bool
}

// logFieldLines gives the field panel of logs.md rule 4 for the raw JSON of
// an event: label and value rows with values cut to valueCells cells per
// row. A nested object indents by two cells and has a blank row before and
// after it. A value that wraps has a blank row after it. A truncated event
// ends with "body cut at 4 KiB" in accent.
func logFieldLines(raw []byte, valueCells int) []string {
	fields, ok := readLogFields(raw)
	if !ok {
		return []string{dimStyle.Render("no fields")}
	}
	slices.SortStableFunc(fields, func(a, b logFieldNode) int {
		return firstFieldRank(a.key) - firstFieldRank(b.key)
	})
	var lines []string
	blank := func() {
		if len(lines) > 0 && lines[len(lines)-1] != "" {
			lines = append(lines, "")
		}
	}
	var add func(nodes []logFieldNode, indent string)
	add = func(nodes []logFieldNode, indent string) {
		for _, node := range nodes {
			label := padCell(dimStyle.Render(indent+node.key), logLabelCells)
			if node.object {
				blank()
				lines = append(lines, dimStyle.Render(indent+node.key))
				add(node.nested, indent+"  ")
				blank()
				continue
			}
			rows := wrapLogValue(sanitizeText(node.value), valueCells)
			style := logValueStyle(node.value)
			for index, row := range rows {
				if index > 0 {
					label = strings.Repeat(" ", logLabelCells)
				}
				lines = append(lines, label+style.Render(row))
			}
			if len(rows) > 1 {
				blank()
			}
		}
	}
	add(fields, "")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, node := range fields {
		if node.key == "truncated" && node.value == "true" {
			lines = append(lines, accentTextStyle.Render("body cut at 4 KiB"))
		}
	}
	return lines
}

// firstFieldRank gives the place of key in logFirstFields, or a place after
// them.
func firstFieldRank(key string) int {
	if index := slices.Index(logFirstFields, key); index >= 0 {
		return index
	}
	return len(logFirstFields)
}

// logValueStyle gives the style of a value: fail and WARN are accent, fail
// in bold, and ERROR is error.
func logValueStyle(value string) lipgloss.Style {
	switch value {
	case "fail":
		return titleStyle
	case "WARN":
		return accentTextStyle
	case "ERROR":
		return errorTextStyle
	}
	return lipgloss.NewStyle()
}

// wrapLogValue wraps value at a space into rows of at most cells cells. A
// word longer than a row breaks at the row width. A rune of more than one
// cell stays whole, so its row has fewer cells than cells.
func wrapLogValue(value string, cells int) []string {
	cells = max(cells, 1)
	var rows []string
	row := ""
	for word := range strings.SplitSeq(value, " ") {
		switch {
		case row == "":
			row = word
		case lipgloss.Width(row)+1+lipgloss.Width(word) <= cells:
			row += " " + word
		default:
			rows = append(rows, row)
			row = word
		}
		rows, row = breakLogRow(rows, row, cells)
	}
	return append(rows, row)
}

// breakLogRow moves each full row of row to rows and gives the rest. A rune
// that is wider than the row goes on a row alone.
func breakLogRow(rows []string, row string, cells int) ([]string, string) {
	for lipgloss.Width(row) > cells {
		runes := []rune(row)
		width, split := 0, 0
		for split < len(runes) {
			next := lipgloss.Width(string(runes[split]))
			if split > 0 && width+next > cells {
				break
			}
			width += next
			split++
		}
		rows = append(rows, string(runes[:split]))
		row = string(runes[split:])
	}
	return rows, row
}

// readLogFields reads the fields of a JSON object in their order.
func readLogFields(raw []byte) ([]logFieldNode, bool) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, false
	}
	fields, err := readLogObject(decoder)
	return fields, err == nil
}

// readLogObject reads the fields of an object after its opening brace. An
// array value shows as its compact JSON.
func readLogObject(decoder *json.Decoder) ([]logFieldNode, error) {
	var fields []logFieldNode
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, _ := token.(string)
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		node := logFieldNode{key: key}
		trimmed := bytes.TrimSpace(value)
		switch {
		case len(trimmed) > 0 && trimmed[0] == '{':
			nested, ok := readLogFields(trimmed)
			if !ok {
				return nil, errBadLogObject
			}
			node.object, node.nested = true, nested
		case len(trimmed) > 0 && trimmed[0] == '"':
			_ = json.Unmarshal(trimmed, &node.value)
		default:
			var compact bytes.Buffer
			_ = json.Compact(&compact, trimmed)
			node.value = compact.String()
		}
		fields = append(fields, node)
	}
	_, err := decoder.Token()
	return fields, err
}
