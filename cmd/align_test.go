package cmd

import (
	"bytes"
	"testing"
)

func TestAlign(t *testing.T) {
	tests := []struct {
		in, out string
		header  bool
	}{
		{"ab\nab", "ab\n||\nab\n", false},
		{"h\nab\nab", "h\nab\n||\nab\n", true},
		{"h\nab\nab\nh\na\nb\n", "h\nab\n||\nab\nh\na\n#\nb\n", true},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			var out bytes.Buffer
			alignArgs.header = tc.header
			if err := align(bytes.NewBufferString(tc.in), &out); err != nil {
				t.Fatalf("got error: %v", err)
			}
			if got := out.String(); got != tc.out {
				t.Fatalf("expected %q; got %q", tc.out, got)
			}
		})
	}
}
