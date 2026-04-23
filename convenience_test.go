package jfmt_test

import (
	"bytes"
	"io"
	"strings"
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

func TestFormatReader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := jfmt.FormatReader(strings.NewReader(`{"b":2,"a":1}`), &buf, jfmt.Options{Indent: "  ", NoFix: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "{\n  \"b\": 2,\n  \"a\": 1\n}"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatReader_compact(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := jfmt.FormatReader(strings.NewReader("{\n  \"a\": 1\n}"), &buf, jfmt.Options{Compact: true, NoFix: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := buf.String(); got != `{"a":1}` {
		t.Errorf("got %q, want %q", got, `{"a":1}`)
	}
}

func TestFormatReader_invalid(t *testing.T) {
	t.Parallel()

	err := jfmt.FormatReader(strings.NewReader(`{bad}`), io.Discard, jfmt.Options{NoFix: true})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFormatReader_readError(t *testing.T) {
	t.Parallel()

	err := jfmt.FormatReader(errReader{}, io.Discard, jfmt.Options{})
	if err == nil {
		t.Fatal("expected error from reader")
	}
}

func TestFixReader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := jfmt.FixReader(strings.NewReader(`{'key':'value'}`), &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := buf.String(); got != `{"key":"value"}` {
		t.Errorf("got %q, want %q", got, `{"key":"value"}`)
	}
}

func TestFixReader_readError(t *testing.T) {
	t.Parallel()

	err := jfmt.FixReader(errReader{}, io.Discard)
	if err == nil {
		t.Fatal("expected error from reader")
	}
}

// errReader always returns an error on Read.
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) { return 0, io.ErrUnexpectedEOF }
