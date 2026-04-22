package main

import (
	"bytes"
	"os"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiGreen  = "\x1b[32m"
	ansiCyan   = "\x1b[36m"
	ansiYellow = "\x1b[33m"
)

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode() & os.ModeCharDevice) != 0
}

// colorize applies ANSI syntax highlighting to formatted JSON.
// Keys are bold, strings are green, numbers are cyan, booleans/null are yellow.
func colorize(src []byte) []byte {
	var (
		out       bytes.Buffer
		stack     []byte // 'o' = object, 'a' = array
		expectKey bool
	)

	i := 0

	for i < len(src) {
		c := src[i]

		switch {
		case c == '{':
			stack = append(stack, 'o')
			expectKey = true
			out.WriteByte(c)
			i++

		case c == '[':
			stack = append(stack, 'a')
			expectKey = false
			out.WriteByte(c)
			i++

		case c == '}' || c == ']':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

			expectKey = false
			out.WriteByte(c)
			i++

		case c == ',':
			if len(stack) > 0 && stack[len(stack)-1] == 'o' {
				expectKey = true
			}

			out.WriteByte(c)
			i++

		case c == ':':
			expectKey = false
			out.WriteByte(c)
			i++

		case c == '"':
			end := scanStringEnd(src, i+1)

			if expectKey {
				out.WriteString(ansiBold)
				expectKey = false
			} else {
				out.WriteString(ansiGreen)
			}

			out.Write(src[i : end+1])
			out.WriteString(ansiReset)
			i = end + 1

		case c == 't' || c == 'f' || c == 'n':
			if kw, n := matchKeyword(src, i); kw != "" {
				out.WriteString(ansiYellow)
				out.WriteString(kw)
				out.WriteString(ansiReset)
				i += n
			} else {
				out.WriteByte(c)
				i++
			}

		case isDigit(c) || c == '-':
			end := i + 1
			for end < len(src) && (isDigit(src[end]) || src[end] == '.' || src[end] == 'e' || src[end] == 'E' || src[end] == '+' || src[end] == '-') {
				end++
			}

			out.WriteString(ansiCyan)
			out.Write(src[i:end])
			out.WriteString(ansiReset)
			i = end

		default:
			out.WriteByte(c)
			i++
		}
	}

	return out.Bytes()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// scanStringEnd returns the index of the closing '"', or len(src)-1 if unterminated.
func scanStringEnd(src []byte, i int) int {
	for i < len(src) {
		if src[i] == '\\' {
			i += 2

			continue
		}

		if src[i] == '"' {
			return i
		}

		i++
	}

	if len(src) == 0 {
		return 0
	}

	return len(src) - 1
}

func matchKeyword(src []byte, i int) (string, int) {
	for _, kw := range []string{"true", "false", "null"} {
		end := i + len(kw)
		if end <= len(src) && string(src[i:end]) == kw {
			if end == len(src) || !isAlphaNumByte(src[end]) {
				return kw, len(kw)
			}
		}
	}

	return "", 0
}

func isAlphaNumByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
