package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/krzysztofgb/jfmt"
)

func applyTemplate(name string, opts *jfmt.Options) bool {
	switch name {
	case "fourspace":
		opts.Indent = "    "
	case "threespace":
		opts.Indent = "   "
	case "twospace":
		opts.Indent = "  "
	case "onetab":
		opts.Indent = "\t"
	case "compact":
		opts.Compact = true
	default:
		return false
	}

	return true
}

func parseSpec(s string) (jfmt.Spec, bool) {
	switch s {
	case "rfc8259":
		return jfmt.RFC8259, true
	case "rfc7159":
		return jfmt.RFC7159, true
	case "rfc4627":
		return jfmt.RFC4627, true
	case "ecma404":
		return jfmt.ECMA404, true
	case "skip":
		return jfmt.SpecSkip, true
	}

	return 0, false
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("jfmt", flag.ContinueOnError)
	flags.SetOutput(stderr)

	template := flags.String("template", "twospace", "indent template: fourspace, threespace, twospace, onetab, compact")
	flags.StringVar(template, "t", "twospace", "shorthand for -template")
	indent := flags.String("indent", "", "custom indent string (overrides -template)")
	flags.StringVar(indent, "i", "", "shorthand for -indent")
	compact := flags.Bool("compact", false, "compact output (overrides -template and -indent)")
	flags.BoolVar(compact, "c", false, "shorthand for -compact")
	specStr := flags.String("spec", "rfc8259", "json spec: rfc8259, rfc7159, rfc4627, ecma404, skip")
	flags.StringVar(specStr, "s", "rfc8259", "shorthand for -spec")
	validateOnly := flags.Bool("validate", false, "validate only, no output")
	flags.BoolVar(validateOnly, "v", false, "shorthand for -validate")
	noFix := flags.Bool("no-fix", false, "disable automatic JSON repair")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	spec, ok := parseSpec(*specStr)
	if !ok {
		fmt.Fprintf(stderr, "jfmt: unknown spec %q (rfc8259, rfc7159, rfc4627, ecma404, skip)\n", *specStr)

		return 1
	}

	var src []byte

	switch flags.NArg() {
	case 0:
		var err error

		src, err = io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: read stdin: %v\n", err)

			return 1
		}

	case 1:
		var err error

		src, err = os.ReadFile(flags.Arg(0))
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: %v\n", err)

			return 1
		}

	default:
		fmt.Fprintln(stderr, "usage: jfmt [flags] [file]")

		return 1
	}

	if *validateOnly {
		if !jfmt.Validate(src, spec) {
			fmt.Fprintln(stderr, "jfmt: invalid JSON")

			return 1
		}

		return 0
	}

	opts := jfmt.Options{
		Spec:  spec,
		NoFix: *noFix,
	}

	switch {
	case *compact:
		opts.Compact = true
	case *indent != "":
		opts.Indent = *indent
	default:
		if !applyTemplate(*template, &opts) {
			fmt.Fprintf(stderr, "jfmt: unknown template %q (fourspace, threespace, twospace, onetab, compact)\n", *template)

			return 1
		}
	}

	out, err := jfmt.Format(src, opts)
	if err != nil {
		fmt.Fprintf(stderr, "jfmt: %v\n", err)

		return 1
	}

	fmt.Fprintln(stdout, string(out))

	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
