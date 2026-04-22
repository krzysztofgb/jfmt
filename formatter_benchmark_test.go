package jfmt_test

import (
	"strings"
	"testing"

	"github.com/krzysztofgb/jfmt"
)

var benchCases = []struct { //nolint:gochecknoglobals
	name  string
	input []byte
}{
	{"small_object", []byte(`{"name":"Alice","age":30,"active":true}`)},
	{"small_array", []byte(`[1,2,3,4,5,6,7,8,9,10]`)},
	{"nested", []byte(`{"user":{"name":"Alice","address":{"city":"Warsaw","country":"PL"}},"tags":["go","json"]}`)},
	{"large_array", buildLargeArray(1000)},
	{"large_object", buildLargeObject(1000)},
}

func buildLargeArray(n int) []byte {
	var b strings.Builder

	b.WriteByte('[')

	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}

		b.WriteString(`{"id":`)
		b.WriteString(strings.Repeat("1", 6))
		b.WriteString(`,"value":"item"}`)
	}

	b.WriteByte(']')

	return []byte(b.String())
}

func buildLargeObject(n int) []byte {
	var b strings.Builder

	b.WriteByte('{')

	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}

		b.WriteString(`"key`)
		b.WriteString(strings.Repeat("0", 4))
		b.WriteString(`":"value`)
		b.WriteString(strings.Repeat("0", 4))
		b.WriteByte('"')
	}

	b.WriteByte('}')

	return []byte(b.String())
}

func BenchmarkFormat_pretty(b *testing.B) {
	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()

			for range b.N {
				_, _ = jfmt.Format(bc.input, jfmt.Options{Indent: "  ", NoFix: true})
			}
		})
	}
}

func BenchmarkFormat_compact(b *testing.B) {
	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()

			for range b.N {
				_, _ = jfmt.Format(bc.input, jfmt.Options{Compact: true, NoFix: true})
			}
		})
	}
}

func BenchmarkValidate(b *testing.B) {
	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()

			for range b.N {
				_ = jfmt.Validate(bc.input, jfmt.RFC8259)
			}
		})
	}
}
