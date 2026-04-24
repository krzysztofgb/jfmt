package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/krzysztofgb/jfmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func formatBytes(src []byte, name string, opts jfmt.Options, verbose bool, stderr io.Writer) ([]byte, error) {
	src = applyFix(src, name, verbose, &opts, stderr)

	return jfmt.Format(src, opts)
}

func processInput(src []byte, name string, opts jfmt.Options, validateOnly bool, verbose bool, useColor bool, spec jfmt.Spec, stdout io.Writer, stderr io.Writer) int {
	if validateOnly {
		if err := jfmt.ValidateError(src, spec); err != nil {
			fmt.Fprintf(stderr, "jfmt: %s: %v\n", name, err)

			return 1
		}

		return 0
	}

	out, err := formatBytes(src, name, opts, verbose, stderr)
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

func processCheckDiff(src []byte, name string, opts jfmt.Options, verbose bool, check, diff, quiet bool, useColor bool, stdout, stderr io.Writer) int {
	formatted, err := formatBytes(src, name, opts, verbose, stderr)
	if err != nil {
		return 1
	}

	if bytes.Equal(src, formatted) {
		return 0
	}

	if diff {
		d := computeDiff(name, src, formatted)
		if useColor {
			d = colorizeDiff(d)
		}

		fmt.Fprint(stdout, d)
	}

	if check {
		if !quiet {
			fmt.Fprintf(stderr, "jfmt: %s: not formatted\n", name)
		}

		return 2
	}

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
	out, err := formatBytes(src, path, opts, verbose, stderr)
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

func expandArgs(args []string, stderr io.Writer) []string {
	var expanded []string

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: %v\n", err)

			continue
		}

		if !info.IsDir() {
			expanded = append(expanded, arg)

			continue
		}

		err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && filepath.Ext(path) == ".json" {
				expanded = append(expanded, path)
			}

			return nil
		})
		if err != nil {
			fmt.Fprintf(stderr, "jfmt: %v\n", err)
		}
	}

	return expanded
}

func buildCmd(exitCode *int, stdout io.Writer, stderr io.Writer) *cobra.Command {
	// cfg holds all configurable options; flags point directly into its fields
	// so applyConfig can merge file config into it without reconstruction.
	var cfg config

	var (
		showVersion   bool
		printCfg      bool
		noConfig      bool
		configFile    string
		quiet         bool
		check         bool
		diff          bool
		recursive     bool
		stdinFilename string
		validateOnly  bool
		write         bool
	)

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:           "jfmt [flags] [file...]",
		Short:         "A JSON formatter and validator for the command line",
		SilenceUsage:  true,
		SilenceErrors: true,
		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return []string{"json"}, cobra.ShellCompDirectiveFilterFileExt
		},
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if noConfig {
				return nil
			}

			fileCfg, err := loadConfig(configFile)
			if err != nil {
				fmt.Fprintf(stderr, "jfmt: config: %v\n", err)

				return nil
			}

			applyConfig(cmd, fileCfg, &cfg)

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Fprintf(stdout, "jfmt %s\n", version)

				return nil
			}

			if printCfg {
				if err := cfg.writeTo(stdout); err != nil {
					fmt.Fprintf(stderr, "jfmt: print-config: %v\n", err)
					*exitCode = 1
				}

				return nil
			}

			if write && (check || diff) {
				fmt.Fprintln(stderr, "jfmt: --write cannot be combined with --check or --diff")
				*exitCode = 1

				return nil
			}

			if validateOnly && (check || diff) {
				fmt.Fprintln(stderr, "jfmt: --validate cannot be combined with --check or --diff")
				*exitCode = 1

				return nil
			}

			spec, ok := parseSpec(cfg.Spec)
			if !ok {
				fmt.Fprintf(stderr, "jfmt: unknown spec %q (rfc8259, rfc7159, rfc4627, ecma404, skip)\n", cfg.Spec)
				*exitCode = 1

				return nil
			}

			opts := jfmt.Options{
				Spec:     spec,
				NoFix:    cfg.NoFix,
				SortKeys: cfg.SortKeys,
			}

			switch {
			case cfg.Compact:
				opts.Compact = true
			case cfg.Indent != "":
				opts.Indent = cfg.Indent
			default:
				if !applyTemplate(cfg.Template, &opts) {
					fmt.Fprintf(stderr, "jfmt: unknown template %q (fourspace, threespace, twospace, onetab, compact)\n", cfg.Template)
					*exitCode = 1

					return nil
				}
			}

			outFile, _ := stdout.(*os.File)
			useColor := !cfg.NoColor && (cfg.Color || (outFile != nil && isTerminal(outFile)))

			if len(args) == 0 {
				if f, ok := cmd.InOrStdin().(*os.File); ok && isTerminal(f) {
					return cmd.Help()
				}

				if write {
					fmt.Fprintln(stderr, "jfmt: --write requires at least one file argument")
					*exitCode = 1

					return nil
				}

				src, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					fmt.Fprintf(stderr, "jfmt: read stdin: %v\n", err)
					*exitCode = 1

					return nil
				}

				name := stdinFilename
				if name == "" {
					name = "<stdin>"
				}

				switch {
				case check || diff:
					*exitCode = processCheckDiff(src, name, opts, cfg.Verbose, check, diff, quiet, useColor, stdout, stderr)
				case cfg.JSONLines:
					*exitCode = processJSONLines(src, name, opts, cfg.Verbose, useColor, stdout, stderr)
				default:
					*exitCode = processInput(src, name, opts, validateOnly, cfg.Verbose, useColor, spec, stdout, stderr)
				}

				return nil
			}

			if recursive {
				args = expandArgs(args, stderr)
			}

			code := 0

			for _, path := range args {
				src, err := os.ReadFile(path)
				if err != nil {
					fmt.Fprintf(stderr, "jfmt: %v\n", err)
					code = 1

					continue
				}

				switch {
				case check || diff:
					if c := processCheckDiff(src, path, opts, cfg.Verbose, check, diff, quiet, useColor, stdout, stderr); c == 1 {
						code = 1
					} else if c == 2 && code != 1 {
						code = 2
					}
				case write:
					if writeInPlace(src, path, opts, cfg.Verbose, stderr) != 0 {
						code = 1
					}
				case cfg.JSONLines:
					if processJSONLines(src, path, opts, cfg.Verbose, useColor, stdout, stderr) != 0 {
						code = 1
					}
				default:
					if processInput(src, path, opts, validateOnly, cfg.Verbose, useColor, spec, stdout, stderr) != 0 {
						code = 1
					}
				}
			}

			*exitCode = code

			return nil
		},
	}

	formattingFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	formattingFlags.StringVarP(&cfg.Template, "template", "t", "twospace", "indent template (fourspace, threespace, twospace, onetab, compact)")
	formattingFlags.StringVarP(&cfg.Indent, "indent", "i", "", "custom indent string (overrides --template)")
	formattingFlags.BoolVarP(&cfg.Compact, "compact", "c", false, "compact output (overrides --template and --indent)")
	formattingFlags.BoolVar(&cfg.SortKeys, "sort-keys", false, "sort object keys alphabetically")

	validationFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	validationFlags.StringVarP(&cfg.Spec, "spec", "s", "rfc8259", "JSON spec (rfc8259, rfc7159, rfc4627, ecma404, skip)")
	validationFlags.BoolVarP(&validateOnly, "validate", "v", false, "validate only, no output")

	outputFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	outputFlags.BoolVarP(&write, "write", "w", false, "write result back to source file")
	outputFlags.BoolVar(&cfg.Color, "color", false, "colorize output (default: auto)")
	outputFlags.BoolVar(&cfg.NoColor, "no-color", false, "disable colorized output")

	repairFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	repairFlags.BoolVar(&cfg.Verbose, "verbose", false, "print repair diagnostics to stderr")
	repairFlags.BoolVar(&cfg.NoFix, "no-fix", false, "disable automatic JSON repair")

	ciFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	ciFlags.BoolVar(&check, "check", false, "exit non-zero if any input is not formatted")
	ciFlags.BoolVarP(&diff, "diff", "d", false, "display diff of changes that would be made")
	ciFlags.BoolVarP(&quiet, "quiet", "q", false, "suppress non-error output")
	ciFlags.BoolVarP(&recursive, "recursive", "r", false, "recursively find and process .json files in directories")

	otherFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	otherFlags.BoolVarP(&cfg.JSONLines, "jsonlines", "l", false, "process input as newline-delimited JSON (NDJSON)")
	otherFlags.StringVar(&stdinFilename, "stdin-filename", "", "filename to use in error messages when reading from stdin")
	otherFlags.StringVar(&configFile, "config", "", "path to config file (default: ~/.config/jfmt/config.toml)")
	otherFlags.BoolVar(&noConfig, "no-config", false, "ignore config file")
	otherFlags.BoolVar(&printCfg, "print-config", false, "print effective configuration and exit")
	otherFlags.BoolVarP(&showVersion, "version", "V", false, "print version and exit")

	for _, fs := range []*pflag.FlagSet{formattingFlags, validationFlags, outputFlags, repairFlags, ciFlags, otherFlags} {
		cmd.Flags().AddFlagSet(fs)
	}

	//nolint:errcheck
	cmd.RegisterFlagCompletionFunc("template", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"fourspace", "threespace", "twospace", "onetab", "compact"}, cobra.ShellCompDirectiveNoFileComp
	})
	//nolint:errcheck
	cmd.RegisterFlagCompletionFunc("spec", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"rfc8259", "rfc7159", "rfc4627", "ecma404", "skip"}, cobra.ShellCompDirectiveNoFileComp
	})

	type flagGroup struct {
		title string
		flags *pflag.FlagSet
	}

	groups := []flagGroup{
		{"Formatting", formattingFlags},
		{"Validation", validationFlags},
		{"Output", outputFlags},
		{"Repair", repairFlags},
		{"CI / Scripting", ciFlags},
		{"Other", otherFlags},
	}

	printHelp := func(w io.Writer) {
		fmt.Fprintf(w, "Usage:\n  %s\n\n%s\n\n", cmd.UseLine(), cmd.Short)

		for _, g := range groups {
			fmt.Fprintf(w, "%s:\n%s\n", g.title, g.flags.FlagUsages())
		}

		fmt.Fprintf(w, "  -h, --help   help for %s\n", cmd.Name())
		fmt.Fprintf(w, "\nRun '%s completion --help' for shell tab-completion setup.\n", cmd.Name())
	}

	defaultHelp := cmd.HelpFunc()
	defaultUsage := cmd.UsageFunc()

	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		if c != cmd {
			defaultHelp(c, args)

			return
		}
		printHelp(c.OutOrStdout())
	})
	cmd.SetUsageFunc(func(c *cobra.Command) error {
		if c != cmd {
			return defaultUsage(c)
		}
		printHelp(c.OutOrStderr())

		return nil
	})

	return cmd
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	exitCode := 0
	cmd := buildCmd(&exitCode, stdout, stderr)
	cmd.SetArgs(args)
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(stderr, "jfmt: %v\n", err)

		return 1
	}

	return exitCode
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
