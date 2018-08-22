package mets

import (
	"testing"
)

func TestFindFileGroups(t *testing.T) {
	tests := []struct {
		fileGrp string
		n       int
		ok      bool
	}{
		{"NOT-A-VALID-FILEGRP", 0, false},
		{"OCR-D-IMG", 0, true},
		{"OCR-D-GT-ALTO", 1, true},
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

func TestFiles(t *testing.T) {
	tests := []struct {
		fileGrp string
		file    File
	}{
		{"OCR-D-GT-ALTO", File{
			MIMEType: "application/alto+xml",
			ID:       "PAGE_0020_ALTO",
			FLocat: FLocat{
				Type: "URL",
				URL:  "https://github.com/OCR-D/assets/raw/master/data/kant_aufklaerung_1784/alto/kant_aufklaerung_1784_0020.xml",
			},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.fileGrp, func(t *testing.T) {
			m, err := Open("testdata/mets.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			fs, ok := m.FindFileGrp(tc.fileGrp)
			if !ok {
				t.Fatalf("could not find file group %q", tc.fileGrp)
			}
			if got := fs[0]; got != tc.file {
				t.Fatalf("expected %v; got %v", tc.file, got)
			}
		})
	}
}
