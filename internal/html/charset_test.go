package html

import "testing"

func TestDetectCharset(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<html><meta charset="UTF-8">`, "utf-8"},
		{`<html><meta charset=utf-8>`, "utf-8"},
		{`<html><meta charset="ISO-8859-1">`, "iso-8859-1"},
		{`<html><meta http-equiv="Content-Type" content="text/html; charset=windows-1252">`, "windows-1252"},
		{`no charset here`, "utf-8"},
		{string([]byte{0xEF, 0xBB, 0xBF}) + `hello`, "utf-8"}, // UTF-8 BOM
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			got := DetectCharset([]byte(tc.input))
			if got != tc.expected {
				t.Errorf("DetectCharset(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
