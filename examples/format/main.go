package main

import (
	"fmt"
	"log"

	"github.com/krzysztofgb/jfmt"
)

func main() {
	src := []byte(`{"name":"Alice","age":30,"tags":["go","json"]}`)

	pretty, err := jfmt.Format(src, jfmt.Options{Indent: "  ", NoFix: true})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("2-space:")
	fmt.Println(string(pretty))

	fourSpace, err := jfmt.Format(src, jfmt.Options{Indent: "    ", NoFix: true})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("4-space:")
	fmt.Println(string(fourSpace))

	tab, err := jfmt.Format(src, jfmt.Options{Indent: "\t", NoFix: true})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("tab:")
	fmt.Println(string(tab))

	compact, err := jfmt.Format(src, jfmt.Options{Compact: true, NoFix: true})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("compact:")
	fmt.Println(string(compact))
}
