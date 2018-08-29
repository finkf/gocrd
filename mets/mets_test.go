package mets

import (
	"log"
	"testing"
)

func TestFindFileGroups(t *testing.T) {
	tests := []struct {
		fileGrp string
		n       int
	}{
		{"NOT-A-VALID-FILEGRP", 0},
		{"OCR-D-IMG", 0},
		{"OCR-D-GT-ALTO", 1},
	}
	for _, tc := range tests {
		t.Run(tc.fileGrp, func(t *testing.T) {
			m, err := Open("testdata/mets.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			fs := m.FindFileGrp(tc.fileGrp)
			if got := len(fs); got != tc.n {
				for _, f := range fs {
					log.Printf("f: %v", f)
				}
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
			fs := m.FindFileGrp(tc.fileGrp)
			if got := fs[0]; got != tc.file {
				t.Fatalf("expected %v; got %v", tc.file, got)
			}
		})
	}
}
