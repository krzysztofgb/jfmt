package jfmt

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
