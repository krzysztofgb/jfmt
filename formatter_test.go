package jfmt_test

import (
	"testing"

	"github.com/krzysztofgb/jfmt"
)

func TestFormat_pretty(t *testing.T) {
	t.Parallel()

	input := []byte(`{"b":2,"a":1}`)
	got, err := jfmt.Format(input, jfmt.Options{Indent: "  ", NoFix: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "{\n  \"b\": 2,\n  \"a\": 1\n}"

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormat_compact(t *testing.T) {
	t.Parallel()

	input := []byte(`{  "a" :  1  }`)
	got, err := jfmt.Format(input, jfmt.Options{Compact: true, NoFix: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"a":1}`

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormat_invalid(t *testing.T) {
	t.Parallel()

	_, err := jfmt.Format([]byte(`{bad}`), jfmt.Options{NoFix: true})

	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestFormat_spec_rfc4627_rejectsNonObject(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		valid bool
	}{
		{`{"key":"value"}`, true},
		{`[1,2,3]`, true},
		{`"string"`, false},
		{`42`, false},
		{`null`, false},
		{`true`, false},
	}

	for _, tc := range cases {
		_, err := jfmt.Format([]byte(tc.input), jfmt.Options{Spec: jfmt.RFC4627, NoFix: true})
		gotErr := err != nil

		if gotErr == tc.valid {
			t.Errorf("RFC4627 Format(%q): got error=%v, want valid=%v", tc.input, err, tc.valid)
		}
	}
}

func TestFormat_specSkip_noValidation(t *testing.T) {
	t.Parallel()

	// SpecSkip still requires syntactically valid JSON for formatting to succeed,
	// but does not enforce spec-level constraints like RFC4627 root type.
	_, err := jfmt.Format([]byte(`"just a string"`), jfmt.Options{Spec: jfmt.SpecSkip, NoFix: true})
	if err != nil {
		t.Errorf("SpecSkip should not reject a valid JSON string: %v", err)
	}
}

func TestFormat_fix_singleQuotes(t *testing.T) {
	t.Parallel()

	got, err := jfmt.Format([]byte(`{'key':'value'}`), jfmt.Options{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"key":"value"}`

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormat_fix_trailingComma(t *testing.T) {
	t.Parallel()

	got, err := jfmt.Format([]byte(`{"a":1,"b":2,}`), jfmt.Options{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"a":1,"b":2}`

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormat_fix_unquotedKeys(t *testing.T) {
	t.Parallel()

	got, err := jfmt.Format([]byte(`{key: "value"}`), jfmt.Options{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"key":"value"}`

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormat_fix_numericKeys(t *testing.T) {
	t.Parallel()

	got, err := jfmt.Format([]byte(`{1: "one", 2: "two"}`), jfmt.Options{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"1":"one","2":"two"}`

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormat_fix_comments(t *testing.T) {
	t.Parallel()

	input := `{
		"a": 1, // line comment
		"b": /* block comment */ 2
	}`
	got, err := jfmt.Format([]byte(input), jfmt.Options{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{"a":1,"b":2}`

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormat_fix_caseInsensitiveLiterals(t *testing.T) {
	t.Parallel()

	got, err := jfmt.Format([]byte(`[True, False, Null, TRUE, FALSE, NULL]`), jfmt.Options{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `[true,false,null,true,false,null]`

	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		valid bool
	}{
		{`{}`, true},
		{`{"key":"value"}`, true},
		{`[1,2,3]`, true},
		{`null`, true},
		{`{bad}`, false},
		{``, false},
	}

	for _, tc := range cases {
		got := jfmt.Validate([]byte(tc.input), jfmt.RFC8259)

		if got != tc.valid {
			t.Errorf("Validate(%q) = %v, want %v", tc.input, got, tc.valid)
		}
	}
}

func TestValidate_rfc4627(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		valid bool
	}{
		{`{}`, true},
		{`[]`, true},
		{`null`, false},
		{`"string"`, false},
		{`42`, false},
	}

	for _, tc := range cases {
		got := jfmt.Validate([]byte(tc.input), jfmt.RFC4627)

		if got != tc.valid {
			t.Errorf("Validate RFC4627(%q) = %v, want %v", tc.input, got, tc.valid)
		}
	}
}

func TestValidate_specSkip(t *testing.T) {
	t.Parallel()

	// SpecSkip always returns true regardless of content.
	for _, input := range []string{`{bad}`, ``, `null`, `{"ok":1}`} {
		if !jfmt.Validate([]byte(input), jfmt.SpecSkip) {
			t.Errorf("Validate SpecSkip(%q) = false, want true", input)
		}
	}
}

func FuzzFormat(f *testing.F) {
	f.Add([]byte(`{"key":"value"}`), false)
	f.Add([]byte(`[1,2,3]`), false)
	f.Add([]byte(`null`), true)
	f.Add([]byte(`{}`), false)
	f.Add([]byte(`[]`), false)
	f.Add([]byte(`"hello"`), false)
	f.Add([]byte(`42`), false)
	f.Add([]byte(`true`), false)
	f.Add([]byte(`{"nested":{"a":1},"arr":[1,2,3]}`), true)
	f.Add([]byte(`[{"a":1},{"b":2},null,true,false,42,"str"]`), true)
	f.Add([]byte(`{"unicode":"こんにちは","emoji":"🎉"}`), false)
	f.Add([]byte(`{"z":3,"a":1,"m":2}`), true)

	f.Fuzz(func(t *testing.T, data []byte, sortKeys bool) {
		t.Helper()

		out, err := jfmt.Format(data, jfmt.Options{Indent: "  ", NoFix: true, SortKeys: sortKeys})
		if err != nil {
			return
		}

		if !jfmt.Validate(out, jfmt.RFC8259) {
			t.Errorf("Format produced invalid JSON for input %q", data)
		}

		// Format must be idempotent: formatting already-formatted output produces the same result.
		out2, err := jfmt.Format(out, jfmt.Options{Indent: "  ", NoFix: true, SortKeys: sortKeys})
		if err != nil {
			t.Errorf("Format of already-formatted output failed for input %q: %v", data, err)

			return
		}

		if string(out) != string(out2) {
			t.Errorf("Format is not idempotent for input %q:\nfirst:  %q\nsecond: %q", data, out, out2)
		}
	})
}

func FuzzFix(f *testing.F) {
	f.Add([]byte(`{"key":"value"}`))
	f.Add([]byte(`{key: 'value', active: True, tags: ['a', 'b',]}`))
	f.Add([]byte(`{"a": 1, // comment` + "\n" + `"b": 2}`))
	f.Add([]byte(`{1: "one", 2: "two"}`))
	f.Add([]byte(`{'a': 1, 'b': [True, False, Null]}`))
	f.Add([]byte(`{"a": 1,}`))
	f.Add([]byte(`/* block */ {"a": 1}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		t.Helper()

		fixed := jfmt.Fix(data)

		// Fix must be idempotent: applying it twice produces the same result.
		fixed2 := jfmt.Fix(fixed)
		if string(fixed) != string(fixed2) {
			t.Errorf("Fix is not idempotent for input %q:\nfirst:  %q\nsecond: %q", data, fixed, fixed2)
		}
	})
}
