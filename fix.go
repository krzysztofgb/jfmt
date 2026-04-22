package jfmt

import (
	"bytes"
	"fmt"
	"strings"
)

// Fix repairs common JSON errors. It handles single-quoted strings,
// unquoted and numeric object keys, case-incorrect literals (True, False, Null),
// line and block comments, trailing commas, and unescaped control characters.
func Fix(src []byte) []byte {
	f := &fixer{
		src:       src,
		pos:       0,
		out:       bytes.Buffer{},
		stack:     nil,
		expectKey: false,
	}

	return f.run()
}

type fixer struct {
	src       []byte
	pos       int
	out       bytes.Buffer
	stack     []byte // 'o' = object context, 'a' = array context
	expectKey bool
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

	case '\'':
		f.scanSingleString()

	default:
		if f.expectKey {
			f.scanUnquotedKey()
		} else if lit, end := f.matchLiteral(); lit != "" {
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
}

func (f *fixer) skipBlockComment() {
	f.pos += 2

	for f.pos+1 < len(f.src) {
		if f.src[f.pos] == '*' && f.src[f.pos+1] == '/' {
			f.pos += 2

			return
		}

		f.pos++
	}
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
			f.out.WriteString(fmt.Sprintf(`\u%04x`, c))
			f.pos++

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
			f.out.WriteString(fmt.Sprintf(`\u%04x`, c))
			f.pos++

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
	f.out.WriteByte('"')

	for _, b := range key {
		switch {
		case b == '"':
			f.out.WriteString(`\"`)
		case b == '\\':
			f.out.WriteString(`\\`)
		case b < 0x20:
			f.out.WriteString(fmt.Sprintf(`\u%04x`, b))
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

		if !strings.EqualFold(string(f.src[f.pos:end]), lit) {
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

		if isSpace(c) {
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
