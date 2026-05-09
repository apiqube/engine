package template

import (
	"regexp"
	"strings"
)

type methodCall struct {
	name string
	args []string
}

// methodCallRegex matches a trailing .Name(args) on the expression. The Name
// must start with a capital letter to disambiguate from path segments.
var methodCallRegex = regexp.MustCompile(`\.([A-Z][a-zA-Z0-9]*)\(([^)]*)\)$`)

// splitPrimaryAndMethods peels trailing method calls off the expression body.
// Returns the primary expression plus an ordered list of method calls.
func splitPrimaryAndMethods(body string) (string, []methodCall) {
	body = strings.TrimSpace(body)
	var methods []methodCall

	for {
		idx := methodCallRegex.FindStringSubmatchIndex(body)
		if idx == nil {
			break
		}
		name := body[idx[2]:idx[3]]
		args := body[idx[4]:idx[5]]
		methods = append([]methodCall{{name: name, args: parseArgs(args)}}, methods...)
		body = strings.TrimRightFunc(body[:idx[0]], unicodeIsSpace)
	}
	return body, methods
}

// parseArgs splits a comma-separated argument list. Single- and double-quoted
// strings have their quotes stripped. Whitespace is trimmed around each arg.
//
// Limitation: nested commas (inside a regex pattern, for instance) are not
// supported. Pass complex arguments via a saved variable instead.
func parseArgs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := splitTopLevelCommas(s)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, trimQuotes(strings.TrimSpace(p)))
	}
	return out
}

// splitTopLevelCommas splits on commas that are not inside single or double quotes.
func splitTopLevelCommas(s string) []string {
	var out []string
	var buf strings.Builder
	var quote byte
	for i := range len(s) {
		c := s[i]
		switch {
		case quote == 0 && (c == '\'' || c == '"'):
			quote = c
			buf.WriteByte(c)
		case quote != 0 && c == quote:
			quote = 0
			buf.WriteByte(c)
		case quote == 0 && c == ',':
			out = append(out, buf.String())
			buf.Reset()
		default:
			buf.WriteByte(c)
		}
	}
	if buf.Len() > 0 {
		out = append(out, buf.String())
	}
	return out
}

func unicodeIsSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
