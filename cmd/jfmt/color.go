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
	var out bytes.Buffer

	i := 0

	for i < len(src) {
		c := src[i]

		switch {
		case c == '"':
			// Look back to determine if this is a key (preceded by { or , at the same level)
			isKey := isObjectKey(src, i)

			end := scanStringEnd(src, i+1)

			if isKey {
				out.WriteString(ansiBold)
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

	return i
}

func isObjectKey(src []byte, pos int) bool {
	j := pos - 1
	for j >= 0 && (src[j] == ' ' || src[j] == '\t' || src[j] == '\n' || src[j] == '\r') {
		j--
	}

	if j < 0 {
		return false
	}

	return src[j] == '{' || src[j] == ','
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
