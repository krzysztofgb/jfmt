# jfmt

[![Go Reference](https://pkg.go.dev/badge/github.com/krzysztofgb/jfmt.svg)](https://pkg.go.dev/github.com/krzysztofgb/jfmt)

A JSON formatter and validator for the command line, also importable as a Go library.

## Install

Pre-built binaries for Linux, macOS, and Windows are available on the [releases page](https://github.com/krzysztofgb/jfmt/releases), built with [GoReleaser](https://goreleaser.com).

To install from source:

```
go install github.com/krzysztofgb/jfmt/cmd/jfmt@latest
```

## Usage

```
jfmt [flags] [file]
```

Reads from a file or stdin. Writes formatted JSON to stdout.

Repair of common JSON errors is on by default: single-quoted strings,
unquoted and numeric object keys, incorrect literals (True/False/Null),
line and block comments, trailing commas, and unescaped control characters.

```
-template, -t   fourspace | threespace | twospace | onetab | compact (default: twospace)
-indent, -i     custom indent string (overrides -template)
-compact, -c    compact output (overrides -template and -indent)
-spec, -s       rfc8259 | rfc7159 | rfc4627 | ecma404 | skip (default: rfc8259)
-validate, -v   validate only, no output
-no-fix         disable automatic JSON repair
```

```bash
jfmt data.json
cat data.json | jfmt -t fourspace
cat data.json | jfmt -compact
jfmt -s rfc4627 -validate data.json
jfmt -no-fix data.json
```

## Library

```bash
go get github.com/krzysztofgb/jfmt
```

```go
out, err := jfmt.Format(src, jfmt.Options{Indent: "  "})
out, err := jfmt.Format(src, jfmt.Options{Compact: true})
ok := jfmt.Validate(src)
```

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
