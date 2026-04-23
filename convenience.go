package jfmt

import "io"

// FormatReader reads JSON from r, repairs/formats it according to opts, and writes the result to w.
func FormatReader(r io.Reader, w io.Writer, opts Options) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	out, err := Format(src, opts)
	if err != nil {
		return err
	}

	_, err = w.Write(out)

	return err
}

// FixReader reads from r, repairs common JSON errors, and writes the result to w.
func FixReader(r io.Reader, w io.Writer) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	_, err = w.Write(Fix(src))

	return err
}

// FormatString formats a JSON string and returns the result as a string.
func FormatString(src string, opts Options) (string, error) {
	out, err := Format([]byte(src), opts)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// ValidateString reports whether src is valid JSON according to spec.
func ValidateString(src string, spec Spec) bool {
	return Validate([]byte(src), spec)
}

// FixString repairs common JSON errors in src and returns the result as a string.
func FixString(src string) string {
	return string(Fix([]byte(src)))
}
