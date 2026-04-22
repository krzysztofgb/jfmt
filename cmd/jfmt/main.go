package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func applyFix(src []byte, name string, verbose bool, opts *jfmt.Options, stderr io.Writer) []byte {
	if !verbose || opts.NoFix {
		return src
	}

	fixed, report := jfmt.FixWithReport(src)
	if s := report.String(); s != "" {
		fmt.Fprintf(stderr, "jfmt: %s: %s\n", name, s)
	}

	opts.NoFix = true

	return fixed
}

func writeOutput(w io.Writer, b []byte) {
	w.Write(append(b, '\n')) //nolint:errcheck
}

func processInput(src []byte, name string, opts jfmt.Options, validateOnly bool, verbose bool, useColor bool, spec jfmt.Spec, stdout io.Writer, stderr io.Writer) int {
	if validateOnly {
		if err := jfmt.ValidateError(src, spec); err != nil {
			fmt.Fprintf(stderr, "jfmt: %s: %v\n", name, err)

			return 1
		}

		return 0
	}

	src = applyFix(src, name, verbose, &opts, stderr)

	out, err := jfmt.Format(src, opts)
	if err != nil {
		fmt.Fprintf(stderr, "jfmt: %s: %v\n", name, err)

		return 1
	}

	if useColor {
		out = colorize(out)
	}

	writeOutput(stdout, out)

	return 0
}

func processJSONLines(src []byte, name string, opts jfmt.Options, verbose bool, useColor bool, stdout io.Writer, stderr io.Writer) int {
	compact := jfmt.Options{
		Compact: true,
		Spec:    opts.Spec,
		NoFix:   opts.NoFix,
	}

	code := 0

	for i, line := range bytes.Split(src, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		lineName := fmt.Sprintf("%s:%d", name, i+1)

		line = applyFix(line, lineName, verbose, &compact, stderr)

		out, err := jfmt.Format(line, compact)
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: %s: %v\n", lineName, err)
			code = 1

			continue
		}

		if useColor {
			out = colorize(out)
		}

		writeOutput(stdout, out)
	}

	return code
}

func writeInPlace(src []byte, path string, opts jfmt.Options, verbose bool, stderr io.Writer) int {
	src = applyFix(src, path, verbose, &opts, stderr)

	out, err := jfmt.Format(src, opts)
	if err != nil {
		fmt.Fprintf(stderr, "jfmt: %s: %v\n", path, err)

		return 1
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".jfmt-*")
	if err != nil {
		fmt.Fprintf(stderr, "jfmt: %s: %v\n", path, err)

		return 1
	}

	if _, err = tmp.Write(out); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		fmt.Fprintf(stderr, "jfmt: %s: %v\n", path, err)

		return 1
	}

	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		fmt.Fprintf(stderr, "jfmt: %s: %v\n", path, err)

		return 1
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		fmt.Fprintf(stderr, "jfmt: %s: %v\n", path, err)

		return 1
	}

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
	write := flags.BoolP("write", "w", false, "write result back to source file")
	sortKeys := flags.Bool("sort-keys", false, "sort object keys alphabetically")
	verbose := flags.Bool("verbose", false, "print repair diagnostics to stderr")
	jsonLines := flags.BoolP("jsonlines", "l", false, "process input as newline-delimited JSON (NDJSON)")
	color := flags.Bool("color", false, "colorize output (default: auto)")
	noColor := flags.Bool("no-color", false, "disable colorized output")
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
		Spec:     spec,
		NoFix:    *noFix,
		SortKeys: *sortKeys,
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

	outFile, _ := stdout.(*os.File)
	useColor := !*noColor && (*color || (outFile != nil && isTerminal(outFile)))

	files := flags.Args()

	if len(files) == 0 {
		if *write {
			fmt.Fprintln(stderr, "jfmt: --write requires at least one file argument")

			return 1
		}

		src, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: read stdin: %v\n", err)

			return 1
		}

		if *jsonLines {
			return processJSONLines(src, "<stdin>", opts, *verbose, useColor, stdout, stderr)
		}

		return processInput(src, "<stdin>", opts, *validateOnly, *verbose, useColor, spec, stdout, stderr)
	}

	code := 0

	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: %v\n", err)
			code = 1

			continue
		}

		if *write {
			if writeInPlace(src, path, opts, *verbose, stderr) != 0 {
				code = 1
			}
		} else if *jsonLines {
			if processJSONLines(src, path, opts, *verbose, useColor, stdout, stderr) != 0 {
				code = 1
			}
		} else if processInput(src, path, opts, *validateOnly, *verbose, useColor, spec, stdout, stderr) != 0 {
			code = 1
		}
	}

	return code
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
