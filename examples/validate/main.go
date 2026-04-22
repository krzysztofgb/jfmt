package main

import (
	"fmt"

	"github.com/krzysztofgb/jfmt"
)

func main() {
	inputs := []string{
		`{"name":"Alice"}`,
		`[1, 2, 3]`,
		`"just a string"`,
		`42`,
		`{bad json}`,
	}

	fmt.Printf("%-26s  %-10s  %-10s\n", "input", "RFC 8259", "RFC 4627")
	fmt.Println("-------------------------------------------------------")

	for _, input := range inputs {
		rfc8259 := jfmt.Validate([]byte(input), jfmt.RFC8259)
		rfc4627 := jfmt.Validate([]byte(input), jfmt.RFC4627)
		fmt.Printf("%-26s  %-10v  %-10v\n", input, rfc8259, rfc4627)
	}
}
