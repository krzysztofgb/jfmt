package jfmt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// sortKeys unmarshals src and re-marshals it, causing object keys to be sorted.
// Numeric precision beyond float64 is not preserved.
func sortKeys(src []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(src, &v); err != nil {
		return nil, err
	}

	out, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Spec identifies the JSON specification to validate against.
type Spec int

const (
	RFC8259  Spec = iota // RFC 8259, current standard (default)
	RFC7159              // RFC 7159
	RFC4627              // RFC 4627: root value must be an object or array
	ECMA404              // ECMA-404
	SpecSkip             // skip validation, format only
)

// Options controls formatting and validation behavior.
// Fix is enabled by default; set NoFix to disable it.
type Options struct {
	Indent   string // indent string; ignored when Compact is true
	Compact  bool   // produce compact JSON; overrides Indent
	Spec     Spec   // JSON specification to validate against
	NoFix    bool   // disable automatic JSON repair
	SortKeys bool   // sort object keys alphabetically
}

// Format repairs (unless opts.NoFix is set), validates, and formats src.
// Returns an error if src is not valid JSON after any repairs.
func Format(src []byte, opts Options) ([]byte, error) {
	if !opts.NoFix {
		src = Fix(src)
	}

	if opts.Spec != SpecSkip {
		if err := checkSpec(src, opts.Spec); err != nil {
			return nil, err
		}
	}

	if opts.SortKeys {
		var err error

		src, err = sortKeys(src)
		if err != nil {
			return nil, fmt.Errorf("sort keys: %w", err)
		}
	}

	if opts.Compact {
		var buf bytes.Buffer

		if err := json.Compact(&buf, src); err != nil {
			return nil, fmt.Errorf("compact: %w", err)
		}

		return buf.Bytes(), nil
	}

	indent := opts.Indent
	if indent == "" {
		indent = "  "
	}

	var buf bytes.Buffer

	if err := json.Indent(&buf, src, "", indent); err != nil {
		return nil, fmt.Errorf("indent: %w", err)
	}

	return buf.Bytes(), nil
}

// Validate reports whether src is valid JSON under spec.
// SpecSkip always returns true.
func Validate(src []byte, spec Spec) bool {
	return ValidateError(src, spec) == nil
}

// ValidateError returns a descriptive error if src is invalid under spec, or nil if valid.
// SpecSkip always returns nil.
func ValidateError(src []byte, spec Spec) error {
	if spec == SpecSkip {
		return nil
	}

	return checkSpec(src, spec)
}

func syntaxLocation(src []byte) (line, col int) {
	dec := json.NewDecoder(bytes.NewReader(src))
	for {
		_, err := dec.Token()
		if err == nil {
			continue
		}

		var se *json.SyntaxError
		if errors.As(err, &se) {
			off := int(se.Offset)
			if off > 0 {
				off--
			}

			line = 1
			col = 1

			for i := 0; i < off && i < len(src); i++ {
				if src[i] == '\n' {
					line++
					col = 1
				} else {
					col++
				}
			}
		}

		break
	}

	return line, col
}

func checkSpec(src []byte, spec Spec) error {
	if !json.Valid(src) {
		line, col := syntaxLocation(src)

		return fmt.Errorf("invalid JSON at line %d, column %d", line, col)
	}

	if spec == RFC4627 {
		trimmed := bytes.TrimSpace(src)
		if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
			return fmt.Errorf("RFC 4627: root value must be an object or array")
		}
	}

	return nil
}
