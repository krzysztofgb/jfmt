package main

import (
	"fmt"
	"log"

	"github.com/krzysztofgb/jfmt"
)

func main() {
	cases := []struct {
		desc  string
		input string
	}{
		{"single quotes", `{'name': 'Alice'}`},
		{"unquoted keys", `{name: "Alice", age: 30}`},
		{"numeric keys", `{1: "one", 2: "two"}`},
		{"incorrect literals", `[True, False, Null]`},
		{"line comment", `{"a": 1, // comment` + "\n" + `"b": 2}`},
		{"block comment", `{"a": 1, /* comment */ "b": 2}`},
		{"trailing comma", `{"a": 1, "b": 2,}`},
	}

	for _, tc := range cases {
		fixed := jfmt.Fix([]byte(tc.input))
		fmt.Printf("%s\n  before: %s\n   after: %s\n\n", tc.desc, tc.input, string(fixed))
	}

	// Format repairs automatically by default.
	broken := []byte(`{name: 'Alice', active: True, tags: ['go', 'json',]}`)

	out, err := jfmt.Format(broken, jfmt.Options{Indent: "  "})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("format with auto-repair:")
	fmt.Println(string(out))
}
