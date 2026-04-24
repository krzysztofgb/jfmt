package jfmt_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/krzysztofgb/jfmt"
)

func ExampleFormat() {
	out, err := jfmt.Format([]byte(`{"b":2,"a":1}`), jfmt.Options{Indent: "  ", NoFix: true})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
	// Output:
	// {
	//   "b": 2,
	//   "a": 1
	// }
}

func ExampleFormat_compact() {
	out, err := jfmt.Format([]byte(`{"a":1,"b":2}`), jfmt.Options{Compact: true, NoFix: true})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
	// Output:
	// {"a":1,"b":2}
}

func ExampleFormat_sortKeys() {
	out, err := jfmt.Format([]byte(`{"b":2,"a":1}`), jfmt.Options{SortKeys: true, NoFix: true})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
	// Output:
	// {
	//   "a": 1,
	//   "b": 2
	// }
}

func ExampleFix() {
	out := jfmt.Fix([]byte(`{'key': 'value', trailing: True,}`))

	fmt.Println(string(out))
	// Output:
	// {"key": "value", "trailing": true}
}

func ExampleFixWithReport() {
	out, report := jfmt.FixWithReport([]byte(`{'a': 1}`))

	fmt.Println(string(out))
	fmt.Println(report)
	// Output:
	// {"a": 1}
	// repaired: 1 single-quoted string(s)
}

func ExampleValidate() {
	fmt.Println(jfmt.Validate([]byte(`{"a":1}`), jfmt.RFC8259))
	fmt.Println(jfmt.Validate([]byte(`"hello"`), jfmt.RFC4627))
	// Output:
	// true
	// false
}

func ExampleValidateError() {
	fmt.Println(jfmt.ValidateError([]byte(`{"a":1}`), jfmt.RFC8259))
	fmt.Println(jfmt.ValidateError([]byte(`{bad}`), jfmt.RFC8259) != nil)
	// Output:
	// <nil>
	// true
}

func ExampleFormatReader() {
	var buf bytes.Buffer

	err := jfmt.FormatReader(strings.NewReader(`{"b":2,"a":1}`), &buf, jfmt.Options{Indent: "  ", NoFix: true})
	if err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
	// Output:
	// {
	//   "b": 2,
	//   "a": 1
	// }
}

func ExampleFixReader() {
	var buf bytes.Buffer

	if err := jfmt.FixReader(strings.NewReader(`{'a':1}`), &buf); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
	// Output:
	// {"a":1}
}
