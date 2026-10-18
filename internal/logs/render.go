package logs

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// TextLine renders one valid event for a person. It gives the local time with
// milliseconds, the four-letter level, the component, the message, and each
// other field as key=value in key order.
func TextLine(event Event) string {
	var builder strings.Builder
	appendPart := func(value string) {
		if builder.Len() != 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(sanitize(value))
	}
	if when, ok := eventTime(event.Fields); ok {
		appendPart(when.Local().Format("2006-01-02T15:04:05.000-07:00"))
	}
	level := strings.ToUpper(fieldString(event.Fields, "level"))
	if level == "" {
		level = "????"
	}
	if len(level) > 4 {
		level = level[:4]
	}
	appendPart(level)
	appendPart(fieldString(event.Fields, "component"))
	appendPart(sanitize(fieldString(event.Fields, "msg")))

	keys := make([]string, 0, len(event.Fields))
	for key := range event.Fields {
		switch key {
		case "time", "level", "component", "msg":
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		appendPart(key + "=" + formatValue(event.Fields[key]))
	}
	return builder.String()
}

// MalformedText renders a malformed line as ? and its first 200 bytes.
func MalformedText(event Event) string {
	prefix := event.Raw
	if len(prefix) > 200 {
		prefix = prefix[:200]
	}
	if len(prefix) == 0 {
		return "?"
	}
	return "? " + sanitize(string(prefix))
}

func eventTime(fields map[string]any) (time.Time, bool) {
	value, ok := fields["time"]
	if !ok {
		return time.Time{}, false
	}
	text, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func formatValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return sanitize(typed)
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(data)
	}
}

func sanitize(text string) string {
	var builder strings.Builder
	for index := 0; index < len(text); {
		r, size := utf8.DecodeRuneInString(text[index:])
		if r == utf8.RuneError && size == 1 {
			builder.WriteByte(text[index])
			index++
			continue
		}
		switch r {
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
				_, _ = fmt.Fprintf(&builder, `\u%04x`, r)
			} else {
				builder.WriteRune(r)
			}
		}
		index += size
	}
	return builder.String()
}
