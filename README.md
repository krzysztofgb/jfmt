# jfmt

[![Go Reference](https://pkg.go.dev/badge/github.com/krzysztofgb/jfmt.svg)](https://pkg.go.dev/github.com/krzysztofgb/jfmt)

A JSON formatter and validator for the command line, also importable as a Go library.

## Install

**Homebrew (macOS and Linux):**

```
brew install --cask krzysztofgb/tap/jfmt
```

Pre-built binaries for Linux, macOS, and Windows are also available on the [releases page](https://github.com/krzysztofgb/jfmt/releases), built with [GoReleaser](https://goreleaser.com).

To install from source:

```
go install github.com/krzysztofgb/jfmt/cmd/jfmt@latest
```

## Usage

```
jfmt [flags] [file ...]
```

Reads from one or more files or stdin. Writes formatted JSON to stdout.

Repair of common JSON errors is on by default: single-quoted strings,
unquoted and numeric object keys, incorrect literals (True/False/Null),
line and block comments, trailing commas, and unescaped control characters.

```
--template, -t   fourspace | threespace | twospace | onetab | compact (default: twospace)
--indent, -i     custom indent string (overrides --template)
--compact, -c    compact output (overrides --template and --indent)
--spec, -s       rfc8259 | rfc7159 | rfc4627 | ecma404 | skip (default: rfc8259)
--validate, -v   validate only, no output
--write, -w      write result back to source file
--sort-keys      sort object keys alphabetically
--verbose        print repair diagnostics to stderr
--jsonlines, -l  process input as newline-delimited JSON (NDJSON)
--color          force color output
--no-color       disable color output
--no-fix         disable automatic JSON repair
--version, -V    print version and exit
```

```bash
jfmt data.json
jfmt -w data.json
jfmt -w *.json
cat data.json | jfmt -t fourspace
cat data.json | jfmt --compact
jfmt -s rfc4627 --validate data.json
jfmt --no-fix data.json
jfmt -cvs rfc4627 data.json
jfmt --sort-keys data.json
jfmt --verbose data.json
cat ndjson.log | jfmt -l
```

Copy formatted output to the clipboard:

```bash
# macOS
jfmt data.json | pbcopy

# Linux
jfmt data.json | xclip -selection clipboard
jfmt data.json | xsel --clipboard --input

# Windows
jfmt data.json | clip
```

## Library

```bash
go get github.com/krzysztofgb/jfmt
```

```go
// Format (fix is on by default)
out, err := jfmt.Format(src, jfmt.Options{Indent: "  "})
out, err := jfmt.Format(src, jfmt.Options{Compact: true})
out, err := jfmt.Format(src, jfmt.Options{Spec: jfmt.RFC4627})
out, err := jfmt.Format(src, jfmt.Options{SortKeys: true})

// Validate against a spec
ok := jfmt.Validate(src, jfmt.RFC8259)
ok := jfmt.Validate(src, jfmt.RFC4627) // root must be object or array
err := jfmt.ValidateError(src, jfmt.RFC8259) // error includes line and column

// Repair without formatting
fixed := jfmt.Fix(src)
fixed, report := jfmt.FixWithReport(src) // report describes what was changed
```

Runnable examples are in [examples/](examples/).

| example | demonstrates |
|---------|-------------|
| [examples/format](examples/format/main.go) | indent templates, compact output |
| [examples/validate](examples/validate/main.go) | spec validation across RFC 8259 and RFC 4627 |
| [examples/repair](examples/repair/main.go) | Fix and auto-repair in Format |

## Benchmarks

```
BenchmarkFormat_pretty/small_object   366 ns/op     80 B/op   1 allocs/op
BenchmarkFormat_pretty/nested         821 ns/op    192 B/op   1 allocs/op
BenchmarkFormat_pretty/large_array    244 µs/op     64 KB/op  1 allocs/op
BenchmarkFormat_compact/small_object  315 ns/op     64 B/op   1 allocs/op
BenchmarkFormat_compact/nested        682 ns/op     96 B/op   1 allocs/op
BenchmarkFormat_compact/large_array   201 µs/op     32 KB/op  1 allocs/op
BenchmarkValidate/small_object        117 ns/op      0 B/op   0 allocs/op
BenchmarkValidate/nested              263 ns/op      0 B/op   0 allocs/op
BenchmarkValidate/large_array          80 µs/op      0 B/op   0 allocs/op
```

## License

See [LICENSE](LICENSE) file for details.
