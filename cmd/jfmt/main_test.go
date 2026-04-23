package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krzysztofgb/jfmt"
)

func runCmd(t *testing.T, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer
	code = run(args, strings.NewReader(stdin), &outBuf, &errBuf)

	return outBuf.String(), errBuf.String(), code
}

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestRun_formatStdin(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"b":2,"a":1}`)

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	want := "{\n  \"b\": 2,\n  \"a\": 1\n}\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestRun_formatFile(t *testing.T) {
	t.Parallel()

	path := writeTempJSON(t, `{"b":2,"a":1}`)
	out, errOut, code := runCmd(t, "", path)

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	want := "{\n  \"b\": 2,\n  \"a\": 1\n}\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestRun_formatMultipleFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.json")
	p2 := filepath.Join(dir, "b.json")

	if err := os.WriteFile(p1, []byte(`{"a":1}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(p2, []byte(`{"b":2}`), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, code := runCmd(t, "", p1, p2)

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	if !strings.Contains(out, `"a": 1`) || !strings.Contains(out, `"b": 2`) {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRun_compact(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a": 1, "b": 2}`, "--compact")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	want := `{"a":1,"b":2}` + "\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestRun_indent(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a":1}`, "--indent", "\t")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	if !strings.Contains(out, "\t") {
		t.Errorf("expected tab indentation, got: %q", out)
	}
}

func TestRun_template(t *testing.T) {
	t.Parallel()

	cases := []struct {
		template string
		indent   string
	}{
		{"fourspace", "    "},
		{"threespace", "   "},
		{"twospace", "  "},
		{"onetab", "\t"},
	}

	for _, tc := range cases {
		t.Run(tc.template, func(t *testing.T) {
			t.Parallel()

			out, errOut, code := runCmd(t, `{"a":1}`, "--template", tc.template)

			if code != 0 {
				t.Fatalf("exit %d, stderr: %s", code, errOut)
			}

			if !strings.Contains(out, tc.indent) {
				t.Errorf("template %q: expected indent %q in output %q", tc.template, tc.indent, out)
			}
		})
	}
}

func TestRun_templateCompact(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a": 1}`, "--template", "compact")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	want := `{"a":1}` + "\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestRun_sortKeys(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"z":1,"a":2}`, "--sort-keys", "--compact")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	want := `{"a":2,"z":1}` + "\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestRun_noFix_invalid(t *testing.T) {
	t.Parallel()

	_, _, code := runCmd(t, `{'a':1}`, "--no-fix")

	if code == 0 {
		t.Fatal("expected non-zero exit for invalid JSON with --no-fix")
	}
}

func TestRun_noFix_valid(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a":1}`, "--no-fix")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	if !strings.Contains(out, `"a": 1`) {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRun_verbose(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{'a':1}`, "--verbose", "--compact")

	if code != 0 {
		t.Fatalf("unexpected exit %d", code)
	}

	if errOut == "" {
		t.Error("expected verbose repair output on stderr, got nothing")
	}
}

func TestRun_write(t *testing.T) {
	t.Parallel()

	path := writeTempJSON(t, `{"b":2,"a":1}`)
	_, errOut, code := runCmd(t, "", "--write", path)

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := "{\n  \"b\": 2,\n  \"a\": 1\n}"
	if string(got) != want {
		t.Errorf("file content: got %q, want %q", got, want)
	}
}

func TestRun_writeRequiresFile(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{"a":1}`, "--write")

	if code == 0 {
		t.Fatal("expected non-zero exit")
	}

	if !strings.Contains(errOut, "--write requires") {
		t.Errorf("expected --write error, got: %q", errOut)
	}
}

func TestRun_validateValid(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a":1}`, "--validate")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	if out != "" {
		t.Errorf("expected no stdout for --validate, got: %q", out)
	}
}

func TestRun_validateInvalid(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{bad}`, "--validate", "--no-fix")

	if code == 0 {
		t.Fatal("expected non-zero exit for invalid JSON")
	}

	if errOut == "" {
		t.Error("expected error message on stderr")
	}
}

func TestRun_checkFormatted(t *testing.T) {
	t.Parallel()

	_, _, code := runCmd(t, "{\n  \"a\": 1\n}", "--check")

	if code != 0 {
		t.Fatalf("expected exit 0 for already-formatted input, got %d", code)
	}
}

func TestRun_checkUnformatted(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{"a":1}`, "--check")

	if code == 0 {
		t.Fatal("expected non-zero exit for unformatted input")
	}

	if !strings.Contains(errOut, "not formatted") {
		t.Errorf("expected 'not formatted' in stderr, got: %q", errOut)
	}
}

func TestRun_checkQuiet(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{"a":1}`, "--check", "--quiet")

	if code == 0 {
		t.Fatal("expected non-zero exit")
	}

	if errOut != "" {
		t.Errorf("expected no stderr with --quiet, got: %q", errOut)
	}
}

func TestRun_diff(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a":1}`, "--diff")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	if !strings.Contains(out, "---") || !strings.Contains(out, "+++") {
		t.Errorf("expected unified diff output, got: %q", out)
	}
}

func TestRun_diffFormatted(t *testing.T) {
	t.Parallel()

	out, _, code := runCmd(t, "{\n  \"a\": 1\n}", "--diff")

	if code != 0 {
		t.Fatalf("expected exit 0 for already-formatted input, got %d", code)
	}

	if out != "" {
		t.Errorf("expected no diff for already-formatted input, got: %q", out)
	}
}

func TestRun_diffAndCheck(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a":1}`, "--diff", "--check")

	if code == 0 {
		t.Fatal("expected non-zero exit with --check")
	}

	if !strings.Contains(out, "---") {
		t.Errorf("expected diff output, got: %q", out)
	}

	if !strings.Contains(errOut, "not formatted") {
		t.Errorf("expected 'not formatted' in stderr, got: %q", errOut)
	}
}

func TestRun_jsonlines(t *testing.T) {
	t.Parallel()

	input := `{"a":1}` + "\n" + `{"b":2}`
	out, errOut, code := runCmd(t, input, "--jsonlines")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 output lines, got %d: %q", len(lines), out)
	}
}

func TestRun_stdinFilename(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{bad}`, "--stdin-filename", "myfile.json", "--validate", "--no-fix")

	if code == 0 {
		t.Fatal("expected non-zero exit")
	}

	if !strings.Contains(errOut, "myfile.json") {
		t.Errorf("expected stdin-filename in error, got: %q", errOut)
	}
}

func TestRun_version(t *testing.T) {
	t.Parallel()

	out, _, code := runCmd(t, "", "--version")

	if code != 0 {
		t.Fatalf("unexpected exit %d", code)
	}

	if !strings.HasPrefix(out, "jfmt ") {
		t.Errorf("unexpected version output: %q", out)
	}
}

func TestRun_spec(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		spec     string
		wantCode int
	}{
		{"rfc8259 valid", `{"a":1}`, "rfc8259", 0},
		{"rfc4627 object valid", `{"a":1}`, "rfc4627", 0},
		{"rfc4627 scalar invalid", `"hello"`, "rfc4627", 1},
		{"ecma404 any valid", `"hello"`, "ecma404", 0},
		{"skip no validation", `"hello"`, "skip", 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, code := runCmd(t, tc.input, "--spec", tc.spec, "--validate", "--no-fix")

			if code != tc.wantCode {
				t.Errorf("exit %d, want %d", code, tc.wantCode)
			}
		})
	}
}

func TestRun_recursive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")

	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		filepath.Join(dir, "a.json"):   `{"a":1}`,
		filepath.Join(sub, "b.json"):   `{"b":2}`,
		filepath.Join(dir, "skip.txt"): `not json`,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	_, errOut, code := runCmd(t, "", "--recursive", "--check", dir)

	if code == 0 {
		t.Fatal("expected non-zero exit: json files are not formatted")
	}

	if strings.Contains(errOut, "skip.txt") {
		t.Error("expected skip.txt to be ignored by --recursive")
	}
}

func TestRun_fileNotFound(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, "", "/nonexistent/path.json")

	if code == 0 {
		t.Fatal("expected non-zero exit for missing file")
	}

	if errOut == "" {
		t.Error("expected error message")
	}
}

func TestRun_invalidJSON(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{bad}`, "--no-fix")

	if code == 0 {
		t.Fatal("expected non-zero exit for invalid JSON")
	}

	if errOut == "" {
		t.Error("expected error message on stderr")
	}
}

func TestRun_incompatibleFlags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args []string
	}{
		{"write and check", []string{"--write", "--check", "/dev/null"}},
		{"write and diff", []string{"--write", "--diff", "/dev/null"}},
		{"validate and check", []string{"--validate", "--check"}},
		{"validate and diff", []string{"--validate", "--diff"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, errOut, code := runCmd(t, `{"a":1}`, tc.args...)

			if code == 0 {
				t.Fatal("expected non-zero exit for incompatible flags")
			}

			if errOut == "" {
				t.Error("expected error message")
			}
		})
	}
}

func TestRun_unknownTemplate(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{"a":1}`, "--template", "nope")

	if code == 0 {
		t.Fatal("expected non-zero exit for unknown template")
	}

	if !strings.Contains(errOut, "nope") {
		t.Errorf("expected template name in error, got: %q", errOut)
	}
}

func TestRun_unknownSpec(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{"a":1}`, "--spec", "nope")

	if code == 0 {
		t.Fatal("expected non-zero exit for unknown spec")
	}

	if !strings.Contains(errOut, "nope") {
		t.Errorf("expected spec name in error, got: %q", errOut)
	}
}

func TestRun_noConfig(t *testing.T) {
	t.Parallel()

	out, errOut, code := runCmd(t, `{"a":1}`, "--no-config")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	if !strings.Contains(out, `"a": 1`) {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestRun_configExplicit(t *testing.T) {
	t.Parallel()

	cfgPath := filepath.Join(t.TempDir(), "config.toml")

	if err := os.WriteFile(cfgPath, []byte(`sort_keys = true`), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, code := runCmd(t, `{"z":1,"a":2}`, "--config", cfgPath, "--compact")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	want := `{"a":2,"z":1}` + "\n"
	if out != want {
		t.Errorf("got %q, want %q (config sort_keys should apply)", out, want)
	}
}

func TestRun_configExplicitMissing(t *testing.T) {
	t.Parallel()

	_, errOut, code := runCmd(t, `{"a":1}`, "--config", "/nonexistent/config.toml")

	if code != 0 {
		t.Fatalf("config error should be a warning, not fatal: exit %d", code)
	}

	if !strings.Contains(errOut, "config") {
		t.Errorf("expected config error on stderr, got: %q", errOut)
	}
}

func TestApplyTemplate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		wantOK  bool
		compact bool
		indent  string
	}{
		{"fourspace", true, false, "    "},
		{"threespace", true, false, "   "},
		{"twospace", true, false, "  "},
		{"onetab", true, false, "\t"},
		{"compact", true, true, ""},
		{"unknown", false, false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var opts jfmt.Options
			ok := applyTemplate(tc.name, &opts)

			if ok != tc.wantOK {
				t.Errorf("applyTemplate(%q) = %v, want %v", tc.name, ok, tc.wantOK)
			}

			if tc.wantOK && opts.Indent != tc.indent {
				t.Errorf("indent: got %q, want %q", opts.Indent, tc.indent)
			}

			if tc.compact && !opts.Compact {
				t.Error("expected Compact=true for 'compact' template")
			}
		})
	}
}

func TestParseSpec(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  string
		wantOK bool
	}{
		{"rfc8259", true},
		{"rfc7159", true},
		{"rfc4627", true},
		{"ecma404", true},
		{"skip", true},
		{"unknown", false},
		{"", false},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			_, ok := parseSpec(tc.input)

			if ok != tc.wantOK {
				t.Errorf("parseSpec(%q) ok=%v, want %v", tc.input, ok, tc.wantOK)
			}
		})
	}
}

func TestExpandArgs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")

	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}

	jsonFile := filepath.Join(dir, "a.json")
	txtFile := filepath.Join(dir, "b.txt")
	nestedJSON := filepath.Join(sub, "c.json")

	for _, f := range []string{jsonFile, txtFile, nestedJSON} {
		if err := os.WriteFile(f, []byte(`{}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var errBuf bytes.Buffer
	result := expandArgs([]string{dir}, &errBuf)

	if errBuf.Len() > 0 {
		t.Fatalf("unexpected error: %s", errBuf.String())
	}

	got := make(map[string]bool, len(result))
	for _, p := range result {
		got[p] = true
	}

	if !got[jsonFile] {
		t.Errorf("expected %s in result", jsonFile)
	}

	if !got[nestedJSON] {
		t.Errorf("expected %s in result", nestedJSON)
	}

	if got[txtFile] {
		t.Errorf("expected %s to be excluded", txtFile)
	}
}

func TestExpandArgs_file(t *testing.T) {
	t.Parallel()

	path := writeTempJSON(t, `{}`)

	var errBuf bytes.Buffer
	result := expandArgs([]string{path}, &errBuf)

	if len(result) != 1 || result[0] != path {
		t.Errorf("expandArgs with file: got %v, want [%s]", result, path)
	}
}

func TestExpandArgs_missing(t *testing.T) {
	t.Parallel()

	var errBuf bytes.Buffer
	result := expandArgs([]string{"/nonexistent"}, &errBuf)

	if len(result) != 0 {
		t.Errorf("expected empty result for missing path, got %v", result)
	}

	if errBuf.Len() == 0 {
		t.Error("expected error message for missing path")
	}
}

func TestComputeDiff(t *testing.T) {
	t.Parallel()

	original := []byte(`{"a":1}`)
	formatted := []byte("{\n  \"a\": 1\n}")
	diff := computeDiff("test.json", original, formatted)

	if !strings.Contains(diff, "---") || !strings.Contains(diff, "+++") {
		t.Errorf("expected unified diff headers, got: %q", diff)
	}
}

func TestColorizeDiff(t *testing.T) {
	t.Parallel()

	diff := "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n+new\n context\n"
	colorized := colorizeDiff(diff)

	if !strings.Contains(colorized, ansiRed) {
		t.Error("expected red color for removed lines")
	}

	if !strings.Contains(colorized, ansiGreen) {
		t.Error("expected green color for added lines")
	}

	if !strings.Contains(colorized, ansiCyan) {
		t.Error("expected cyan color for hunk headers")
	}
}

func TestLoadConfig_missing(t *testing.T) {
	t.Parallel()

	_, err := loadConfig("/nonexistent/path/config.toml")

	if err == nil {
		t.Fatal("expected error for explicit missing config path")
	}
}

func TestLoadConfig_valid(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")

	if err := os.WriteFile(path, []byte("template = \"fourspace\"\nsort_keys = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Template != "fourspace" {
		t.Errorf("template: got %q, want %q", cfg.Template, "fourspace")
	}

	if cfg.SortKeys == nil || !*cfg.SortKeys {
		t.Error("expected sort_keys = true")
	}
}

func TestLoadConfig_invalidTOML(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")

	if err := os.WriteFile(path, []byte("not = [valid toml"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := loadConfig(path)

	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestConfigPath_XDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")

	path := configPath()

	want := filepath.Join("/custom/config", "jfmt", "config.toml")
	if path != want {
		t.Errorf("got %q, want %q", path, want)
	}
}

func TestRun_jsonlines_invalidLine(t *testing.T) {
	t.Parallel()

	input := `{"a":1}` + "\n" + `{bad}`
	out, errOut, code := runCmd(t, input, "--jsonlines", "--no-fix")

	if code == 0 {
		t.Fatal("expected non-zero exit for invalid NDJSON line")
	}

	// Valid line should still be output before the error
	if !strings.Contains(out, `{"a":1}`) {
		t.Errorf("expected valid line in output, got: %q", out)
	}

	if errOut == "" {
		t.Error("expected error message for invalid NDJSON line")
	}
}

func TestApplyConfig_keys(t *testing.T) {
	t.Parallel()

	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	content := "template = \"compact\"\nsort_keys = true\n"

	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, code := runCmd(t, `{"z":1,"a":2}`, "--config", cfgPath)

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	want := `{"a":2,"z":1}` + "\n"
	if out != want {
		t.Errorf("got %q, want %q (config should apply compact+sort_keys)", out, want)
	}
}

func TestApplyConfig_cliOverridesConfig(t *testing.T) {
	t.Parallel()

	cfgPath := filepath.Join(t.TempDir(), "config.toml")

	if err := os.WriteFile(cfgPath, []byte(`template = "compact"`), 0o600); err != nil {
		t.Fatal(err)
	}

	// CLI --template should override config template
	out, errOut, code := runCmd(t, `{"a":1}`, "--config", cfgPath, "--template", "fourspace")

	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut)
	}

	if !strings.Contains(out, "    ") {
		t.Errorf("expected fourspace indent (CLI override), got: %q", out)
	}
}
