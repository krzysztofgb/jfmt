package main

import "testing"

func FuzzColorize(f *testing.F) {
	f.Add([]byte(`{"name":"Alice","count":42,"ok":true,"tags":["go","json"]}`))
	f.Add([]byte(`[1,"two",true,null,{"nested":"value"}]`))
	f.Add([]byte(`"unterminated`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		t.Helper()
		_ = colorize(data)
	})
}
