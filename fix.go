package jfmt

import (
	"bytes"
	"fmt"
	"strings"
)

// FixReport counts repairs made by Fix or FixWithReport.
type FixReport struct {
	SingleQuotes   int // single-quoted strings converted to double-quoted
	UnquotedKeys   int // unquoted object keys quoted
	NumericKeys    int // numeric object keys quoted
	Literals       int // case-incorrect literals corrected (True→true, etc.)
	LineComments   int // line comments removed
	BlockComments  int // block comments removed
	TrailingCommas int // trailing commas removed
	ControlChars   int // unescaped control characters escaped
}

// String returns a human-readable summary of repairs, or empty string if none.
func (r FixReport) String() string {
	type entry struct {
		n    int
		desc string
	}

	entries := []entry{
		{r.SingleQuotes, "single-quoted string(s)"},
		{r.UnquotedKeys, "unquoted key(s)"},
		{r.NumericKeys, "numeric key(s)"},
		{r.Literals, "incorrect literal(s)"},
		{r.LineComments, "line comment(s)"},
		{r.BlockComments, "block comment(s)"},
		{r.TrailingCommas, "trailing comma(s)"},
		{r.ControlChars, "unescaped control character(s)"},
	}

	var parts []string

	for _, e := range entries {
		if e.n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", e.n, e.desc))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return "repaired: " + strings.Join(parts, ", ")
}

// Fix repairs common JSON errors. It handles single-quoted strings,
// unquoted and numeric object keys, case-incorrect literals (True, False, Null),
// line and block comments, trailing commas, and unescaped control characters.
func Fix(src []byte) []byte {
	out, _ := FixWithReport(src)

	return out
}

// FixWithReport repairs src and returns a report of what was changed.
func FixWithReport(src []byte) ([]byte, FixReport) {
	f := &fixer{src: src}

	return f.run(), f.report
}

type fixer struct {
	src       []byte
	pos       int
	out       bytes.Buffer
	stack     []byte // 'o' = object context, 'a' = array context
	expectKey bool
	report    FixReport
}

func (f *fixer) run() []byte {
	for f.pos < len(f.src) {
		f.step()
	}

	return f.out.Bytes()
}

func (f *fixer) inObj() bool {
	return len(f.stack) > 0 && f.stack[len(f.stack)-1] == 'o'
}

func (f *fixer) step() {
	c := f.src[f.pos]

	if c == '/' && f.pos+1 < len(f.src) {
		if f.src[f.pos+1] == '/' {
			f.skipLineComment()

			return
		}

		if f.src[f.pos+1] == '*' {
			f.skipBlockComment()

			return
		}
	}

	if isSpace(c) {
		f.out.WriteByte(c)
		f.pos++

		return
	}

	switch c {
	case '{':
		f.stack = append(f.stack, 'o')
		f.out.WriteByte(c)
		f.pos++
		f.expectKey = true

	case '[':
		f.stack = append(f.stack, 'a')
		f.out.WriteByte(c)
		f.pos++
		f.expectKey = false

	case '}', ']':
		if len(f.stack) > 0 {
			f.stack = f.stack[:len(f.stack)-1]
		}

		f.out.WriteByte(c)
		f.pos++
		f.expectKey = false

	case ',':
		if next := f.peekSignificant(f.pos + 1); next == '}' || next == ']' {
			f.pos++ // trailing comma
			f.report.TrailingCommas++
		} else {
			f.out.WriteByte(c)
			f.pos++

			if f.inObj() {
				f.expectKey = true
			}
		}

	case ':':
		f.out.WriteByte(c)
		f.pos++
		f.expectKey = false

	case '"':
		f.scanDoubleString()
		f.expectKey = false

	case '\'':
		f.scanSingleString()
		f.expectKey = false
		f.report.SingleQuotes++

	default:
		if f.expectKey {
			f.scanUnquotedKey()
		} else if lit, end := f.matchLiteral(); lit != "" {
			if string(f.src[f.pos:end]) != lit {
				f.report.Literals++
			}

			f.out.WriteString(lit)
			f.pos = end
		} else {
			f.out.WriteByte(c)
			f.pos++
		}
	}
}

func (f *fixer) skipLineComment() {
	f.pos += 2

	for f.pos < len(f.src) && f.src[f.pos] != '\n' {
		f.pos++
	}

	// Trim whitespace that appeared before the // on the same line.
	b := f.out.Bytes()
	end := len(b)

	for end > 0 && (b[end-1] == ' ' || b[end-1] == '\t') {
		end--
	}

	f.out.Truncate(end)
	f.report.LineComments++
}

func (f *fixer) skipBlockComment() {
	f.pos += 2

	for f.pos+1 < len(f.src) {
		if f.src[f.pos] == '*' && f.src[f.pos+1] == '/' {
			f.pos += 2
			f.report.BlockComments++

			return
		}

		f.pos++
	}

	f.report.BlockComments++ // unterminated: consumed to EOF
}

func (f *fixer) scanDoubleString() {
	f.out.WriteByte('"')
	f.pos++

	for f.pos < len(f.src) {
		c := f.src[f.pos]

		switch {
		case c == '\\':
			f.out.WriteByte(c)
			f.pos++

			if f.pos < len(f.src) {
				f.out.WriteByte(f.src[f.pos])
				f.pos++
			}

		case c == '"':
			f.out.WriteByte(c)
			f.pos++

			return

		case c < 0x20:
			fmt.Fprintf(&f.out, `\u%04x`, c)
			f.pos++
			f.report.ControlChars++

		default:
			f.out.WriteByte(c)
			f.pos++
		}
	}
}

func (f *fixer) scanSingleString() {
	f.out.WriteByte('"')
	f.pos++

	for f.pos < len(f.src) {
		c := f.src[f.pos]

		if c == '\\' && f.pos+1 < len(f.src) {
			next := f.src[f.pos+1]

			switch next {
			case '\'':
				f.out.WriteByte('\'')
				f.pos += 2
			case '"':
				f.out.WriteString(`\"`)
				f.pos += 2
			default:
				f.out.WriteByte(c)
				f.out.WriteByte(next)
				f.pos += 2
			}

			continue
		}

		switch {
		case c == '"':
			f.out.WriteString(`\"`)
			f.pos++

		case c == '\'':
			f.out.WriteByte('"')
			f.pos++

			return

		case c < 0x20:
			fmt.Fprintf(&f.out, `\u%04x`, c)
			f.pos++
			f.report.ControlChars++

		default:
			f.out.WriteByte(c)
			f.pos++
		}
	}

	f.out.WriteByte('"') // unterminated string
}

func (f *fixer) scanUnquotedKey() {
	start := f.pos

	for f.pos < len(f.src) {
		c := f.src[f.pos]
		if c == ':' || c == '}' || c == '\r' || c == '\n' {
			break
		}

		f.pos++
	}

	key := bytes.TrimSpace(f.src[start:f.pos])

	isNumeric := len(key) > 0
	for _, b := range key {
		if b < '0' || b > '9' {
			isNumeric = false

			break
		}
	}

	if isNumeric {
		f.report.NumericKeys++
	} else {
		f.report.UnquotedKeys++
	}

	f.out.WriteByte('"')

	for _, b := range key {
		switch {
		case b == '"':
			f.out.WriteString(`\"`)
		case b == '\\':
			f.out.WriteString(`\\`)
		case b < 0x20:
			fmt.Fprintf(&f.out, `\u%04x`, b)
		default:
			f.out.WriteByte(b)
		}
	}

	f.out.WriteByte('"')
	f.expectKey = false
}

func (f *fixer) matchLiteral() (string, int) {
	for _, lit := range []string{"true", "false", "null"} {
		end := f.pos + len(lit)

		if end > len(f.src) {
			continue
		}

		if !bytes.EqualFold(f.src[f.pos:end], []byte(lit)) {
			continue
		}

		if end < len(f.src) && isAlphaNum(f.src[end]) {
			continue
		}

		return lit, end
	}

	return "", 0
}

func (f *fixer) peekSignificant(pos int) byte {
	for pos < len(f.src) {
		c := f.src[pos]

		if isSpace(c) || c == ',' {
			pos++

			continue
		}

		if c == '/' && pos+1 < len(f.src) {
			if f.src[pos+1] == '/' {
				for pos < len(f.src) && f.src[pos] != '\n' {
					pos++
				}

				continue
			}

			if f.src[pos+1] == '*' {
				pos += 2

				for pos+1 < len(f.src) {
					if f.src[pos] == '*' && f.src[pos+1] == '/' {
						pos += 2

						break
					}

					pos++
				}

				continue
			}
		}

		return c
	}

	return 0
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
