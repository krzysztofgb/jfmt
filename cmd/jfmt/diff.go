package main

import (
	"fmt"
	"strings"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

func computeDiff(name string, original, formatted []byte) string {
	edits := myers.ComputeEdits(span.URI(name), string(original), string(formatted))
	unified := gotextdiff.ToUnified("a/"+name, "b/"+name, string(original), edits)

	return fmt.Sprint(unified)
}

func colorizeDiff(d string) string {
	lines := strings.Split(d, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
			lines[i] = ansiBold + line + ansiReset
		case strings.HasPrefix(line, "@@"):
			lines[i] = ansiCyan + line + ansiReset
		case strings.HasPrefix(line, "-"):
			lines[i] = ansiRed + line + ansiReset
		case strings.HasPrefix(line, "+"):
			lines[i] = ansiGreen + line + ansiReset
		}
	}

	return strings.Join(lines, "\n")
}
