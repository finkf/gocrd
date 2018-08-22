package mets

import (
	"testing"
)

func TestMetsFindFileGroups(t *testing.T) {
	tests := []struct {
		fileGrp string
		n       int
		ok      bool
	}{
		{"NOT-A-VALID-FILEGRP", 0, false},
		{"OCR-D-GT-ALTO", 1, true},
		{"OCR-D-IMG", 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.fileGrp, func(t *testing.T) {
			m, err := Open("testdata/mets.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			fs, ok := m.FindFileGrp(tc.fileGrp)
			if ok != tc.ok {
				t.Fatalf("expected %t; got %t", tc.ok, ok)
			}
			if got := len(fs); got != tc.n {
				t.Fatalf("expected %d; got %d", tc.n, got)
			}
		})
	}
}
