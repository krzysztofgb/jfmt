package jfmt_test

import (
	"testing"

	"github.com/krzysztofgb/jfmt"
)

func TestFix_validJSON_unchanged(t *testing.T) {
	t.Parallel()

	cases := []string{
		`{}`,
		`[]`,
		`{"key":"value"}`,
		`[1,2,3]`,
		`null`,
		`true`,
		`false`,
		`42`,
		`"hello"`,
		`{"nested":{"a":1},"arr":[1,2,3]}`,
	}

	for _, input := range cases {
		got := string(jfmt.Fix([]byte(input)))

		if got != input {
			t.Errorf("Fix(%q) = %q, want unchanged", input, got)
		}
	}
}

func TestFix_singleQuotes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{`{'key':'value'}`, `{"key":"value"}`},
		{`{'a':1,'b':2}`, `{"a":1,"b":2}`},
		{`{'it\'s':'fine'}`, `{"it's":"fine"}`},
		{`{'has"quote':'val'}`, `{"has\"quote":"val"}`},
	}

	for _, tc := range cases {
		got := string(jfmt.Fix([]byte(tc.input)))

		if got != tc.want {
			t.Errorf("Fix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFix_unquotedKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{`{key: "value"}`, `{"key": "value"}`},
		{`{a: 1, b: 2}`, `{"a": 1, "b": 2}`},
		{`{camelCase: true}`, `{"camelCase": true}`},
		{`{snake_case: 1}`, `{"snake_case": 1}`},
	}

	for _, tc := range cases {
		got := string(jfmt.Fix([]byte(tc.input)))

		if got != tc.want {
			t.Errorf("Fix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFix_numericKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{`{1: "one"}`, `{"1": "one"}`},
		{`{42: "forty-two"}`, `{"42": "forty-two"}`},
	}

	for _, tc := range cases {
		got := string(jfmt.Fix([]byte(tc.input)))

		if got != tc.want {
			t.Errorf("Fix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFix_trailingCommas(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{`{"a":1,}`, `{"a":1}`},
		{`[1,2,3,]`, `[1,2,3]`},
		{`{"a":1,"b":[1,2,],}`, `{"a":1,"b":[1,2]}`},
	}

	for _, tc := range cases {
		got := string(jfmt.Fix([]byte(tc.input)))

		if got != tc.want {
			t.Errorf("Fix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFix_lineComments(t *testing.T) {
	t.Parallel()

	input := `{
	"a": 1, // first
	"b": 2  // second
}`
	want := `{
	"a": 1,
	"b": 2
}`

	got := string(jfmt.Fix([]byte(input)))

	if got != want {
		t.Errorf("Fix line comments:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestFix_blockComments(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{`{"a":/* comment */1}`, `{"a":1}`},
		{`[1,/* multi\nline */2]`, `[1,2]`},
	}

	for _, tc := range cases {
		got := string(jfmt.Fix([]byte(tc.input)))

		if got != tc.want {
			t.Errorf("Fix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFix_caseInsensitiveLiterals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{`[True]`, `[true]`},
		{`[False]`, `[false]`},
		{`[Null]`, `[null]`},
		{`[TRUE,FALSE,NULL]`, `[true,false,null]`},
		{`{"ok":True}`, `{"ok":true}`},
	}

	for _, tc := range cases {
		got := string(jfmt.Fix([]byte(tc.input)))

		if got != tc.want {
			t.Errorf("Fix(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
