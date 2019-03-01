package cmd

import (
	"bytes"
	"testing"
)

func TestZip(t *testing.T) {
	tests := []struct {
		in, out, delim string
	}{
		{"a\nb\nc\nd", "a\nc\nb\nd\n", "\n"},
		{"a\nb\nc\nd\ne", "a\nc\nb\nd\n", "\n"},
		{"a\nb\nc\nd", "a-c\nb-d\n", "-"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			zipArgs.delim = tc.delim
			in := bytes.NewBufferString(tc.in)
			var out bytes.Buffer
			if err := zip(in, &out); err != nil {
				t.Fatalf("got error: %v", err)
			}
			if got := out.String(); got != tc.out {
				t.Fatalf("expected %q; got %q", tc.out, got)
			}
		})
	}
}
