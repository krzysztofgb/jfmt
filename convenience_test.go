package jfmt_test

import (
	"testing"

	"github.com/krzysztofgb/jfmt"
)

func TestFormatString(t *testing.T) {
	t.Parallel()

	got, err := jfmt.FormatString(`{"b":2,"a":1}`, jfmt.Options{Indent: "  ", NoFix: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "{\n  \"b\": 2,\n  \"a\": 1\n}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatString_invalid(t *testing.T) {
	t.Parallel()

	_, err := jfmt.FormatString(`{bad}`, jfmt.Options{NoFix: true})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidateString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		spec  jfmt.Spec
		valid bool
	}{
		{`{"a":1}`, jfmt.RFC8259, true},
		{`{bad}`, jfmt.RFC8259, false},
		{`"hello"`, jfmt.RFC4627, false},
		{`{"a":1}`, jfmt.RFC4627, true},
	}

	for _, tc := range cases {
		got := jfmt.ValidateString(tc.input, tc.spec)
		if got != tc.valid {
			t.Errorf("ValidateString(%q, %v) = %v, want %v", tc.input, tc.spec, got, tc.valid)
		}
	}
}

func TestFixString(t *testing.T) {
	t.Parallel()

	got := jfmt.FixString(`{'key':'value'}`)
	want := `{"key":"value"}`

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
