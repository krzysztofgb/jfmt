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

### Shell Completion

After installing, set up tab completion for your shell:

```bash
# Bash
jfmt completion bash >> ~/.bashrc

# Zsh
jfmt completion zsh >> ~/.zshrc

# Fish
jfmt completion fish > ~/.config/fish/completions/jfmt.fish

# PowerShell
jfmt completion powershell >> $PROFILE
```

## Usage

```
jfmt [flags] [file ...]
```

Reads from one or more files or stdin. Writes formatted JSON to stdout.

Repair of common JSON errors is on by default: single-quoted strings,
unquoted and numeric object keys, incorrect literals (True/False/Null),
line and block comments, trailing commas, and unescaped control characters.

**Formatting**

| Flag          | Short | Default    | Description                                                                 |
|---------------|-------|------------|-----------------------------------------------------------------------------|
| `--template`  | `-t`  | `twospace` | Indent template: `fourspace`, `threespace`, `twospace`, `onetab`, `compact` |
| `--indent`    | `-i`  |            | Custom indent string (overrides `--template`)                               |
| `--compact`   | `-c`  |            | Compact output (overrides `--template` and `--indent`)                      |
| `--sort-keys` |       |            | Sort object keys alphabetically                                             |

**Validation**

| Flag         | Short | Default   | Description                                                              |
|--------------|-------|-----------|--------------------------------------------------------------------------|
| `--spec`     | `-s`  | `rfc8259` | JSON spec to validate against: `rfc8259`, `rfc7159`, `rfc4627`, `ecma404`, `skip` |
| `--validate` | `-v`  |           | Validate only, no output                                                 |

**Output**

| Flag        | Short | Description                    |
|-------------|-------|--------------------------------|
| `--write`   | `-w`  | Write result back to source file |
| `--color`   |       | Force color output             |
| `--no-color`|       | Disable color output           |

**Repair**

| Flag        | Description                          |
|-------------|--------------------------------------|
| `--verbose` | Print repair diagnostics to stderr   |
| `--no-fix`  | Disable automatic JSON repair        |

**CI / scripting**

| Flag            | Short | Description                                               |
|-----------------|-------|-----------------------------------------------------------|
| `--check`       |       | Exit non-zero if any input is not formatted (no output)   |
| `--diff`        | `-d`  | Display diff of changes that would be made                |
| `--recursive`   | `-r`  | Recursively find and process `.json` files in directories |
| `--quiet`       | `-q`  | Suppress non-error output (useful with `--check`)         |

**Other**

| Flag               | Short | Description                                               |
|--------------------|-------|-----------------------------------------------------------|
| `--jsonlines`      | `-l`  | Process input as newline-delimited JSON (NDJSON)          |
| `--stdin-filename` |       | Filename for error messages when reading from stdin       |
| `--config`         |       | Path to config file                                       |
| `--no-config`      |       | Ignore config file                                        |
| `--version`        | `-V`  | Print version and exit                                    |

```bash
jfmt data.json
jfmt -w data.json
jfmt -w *.json
jfmt -rw ./data              # format all .json files under ./data
cat data.json | jfmt -t fourspace
cat data.json | jfmt --compact
jfmt -s rfc4627 --validate data.json
jfmt --no-fix data.json
jfmt -cvs rfc4627 data.json
jfmt --sort-keys data.json
jfmt --verbose data.json
cat ndjson.log | jfmt -l
jfmt --check *.json          # CI: fail if any file is not formatted
jfmt --diff data.json        # preview changes before writing
```

## Exit codes

| Code | Meaning                                                    |
|------|------------------------------------------------------------|
| `0`  | Success                                                    |
| `1`  | Error (I/O failure, invalid JSON, invalid flag)            |
| `1`  | `--check`: one or more files are not formatted             |

## Configuration

jfmt looks for a config file at:

- **macOS/Linux**: `$XDG_CONFIG_HOME/jfmt/config.toml` (defaults to `~/.config/jfmt/config.toml`)
- **Windows**: `%APPDATA%\jfmt\config.toml`

A missing config file is silently ignored. Use `--no-config` to bypass it entirely for scripting.

All keys are optional and correspond to their flag equivalents. CLI flags always take priority over config values.

```toml
# ~/.config/jfmt/config.toml

template  = "fourspace"  # fourspace | threespace | twospace | onetab | compact
spec      = "rfc8259"  # rfc8259 | rfc7159 | rfc4627 | ecma404 | skip
sort_keys = false
no_fix    = false
color     = false
no_color  = false
verbose   = false
```

## Library

Use jfmt as a Go library when you need to format, validate, or repair JSON programmatically.

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

| Example                                            | Demonstrates                                      |
|----------------------------------------------------|---------------------------------------------------|
| [examples/format](examples/format/main.go)         | indent templates, compact output                  |
| [examples/validate](examples/validate/main.go)     | spec validation across RFC 8259 and RFC 4627      |
| [examples/repair](examples/repair/main.go)         | Fix and auto-repair in Format                     |

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
