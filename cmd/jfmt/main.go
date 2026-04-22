package main

import (
	"fmt"
	"io"
	"os"

	"github.com/krzysztofgb/jfmt"
	flag "github.com/spf13/pflag"
)

var version = "dev"

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

func processInput(src []byte, name string, opts jfmt.Options, validateOnly bool, spec jfmt.Spec, stdout io.Writer, stderr io.Writer) int {
	if validateOnly {
		if !jfmt.Validate(src, spec) {
			fmt.Fprintf(stderr, "jfmt: %s: invalid JSON\n", name)

			return 1
		}

		return 0
	}

	out, err := jfmt.Format(src, opts)
	if err != nil {
		fmt.Fprintf(stderr, "jfmt: %s: %v\n", name, err)

		return 1
	}

	fmt.Fprintln(stdout, string(out))

	return 0
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("jfmt", flag.ContinueOnError)
	flags.SetOutput(stderr)

	showVersion := flags.BoolP("version", "V", false, "print version and exit")
	template := flags.StringP("template", "t", "twospace", "indent template: fourspace, threespace, twospace, onetab, compact")
	indent := flags.StringP("indent", "i", "", "custom indent string (overrides --template)")
	compact := flags.BoolP("compact", "c", false, "compact output (overrides --template and --indent)")
	specStr := flags.StringP("spec", "s", "rfc8259", "json spec: rfc8259, rfc7159, rfc4627, ecma404, skip")
	validateOnly := flags.BoolP("validate", "v", false, "validate only, no output")
	noFix := flags.Bool("no-fix", false, "disable automatic JSON repair")

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if *showVersion {
		fmt.Fprintf(stdout, "jfmt %s\n", version)
		return 0
	}

	spec, ok := parseSpec(*specStr)
	if !ok {
		fmt.Fprintf(stderr, "jfmt: unknown spec %q (rfc8259, rfc7159, rfc4627, ecma404, skip)\n", *specStr)

		return 1
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

	files := flags.Args()

	if len(files) == 0 {
		src, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: read stdin: %v\n", err)

			return 1
		}

		return processInput(src, "<stdin>", opts, *validateOnly, spec, stdout, stderr)
	}

	code := 0

	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: %v\n", err)
			code = 1

			continue
		}

		if processInput(src, path, opts, *validateOnly, spec, stdout, stderr) != 0 {
			code = 1
		}
	}

	return code
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
