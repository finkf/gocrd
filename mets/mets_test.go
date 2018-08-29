package mets

import (
	"fmt"
	"testing"
)

func TestFindFileGroups(t *testing.T) {
	m, err := Open("testdata/mets.xml")
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	tests := []struct {
		fileGrp string
		n       int
	}{
		{"NOT-A-VALID-FILEGRP", 0},
		{"OCR-D-IMG", 0},
		{"OCR-D-GT-PAGE", 1},
		{"OCR-D-GT-ALTO", 2},
	}
	for _, tc := range tests {
		t.Run(tc.fileGrp, func(t *testing.T) {
			fs := m.FindFileGrp(tc.fileGrp)
			if got := len(fs); got != tc.n {
				t.Fatalf("expected %d; got %d", tc.n, got)
			}
		})
	}
}

func TestFind(t *testing.T) {
	m, err := Open("testdata/mets.xml")
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	tests := []struct {
		m Match
		n int
	}{
		{Match{Use: "OCR-D-IMG"}, 0},
		{Match{Use: "OCR-D-GT-PAGE"}, 1},
		{Match{Use: "OCR-D-GT-ALTO"}, 2},
		{Match{MIMEType: "application/alto+xml"}, 2},
		{Match{MIMEType: "imge/tiff"}, 0},
		{Match{MIMEType: "application/vnd.prima.page+xml"}, 1},
		{Match{FileID: "PAGE_0020_ALTO"}, 1},
		{Match{FileID: "PAGE_0020_PAGE"}, 1},
		{Match{FileID: "PAGE_0021_ALTO"}, 1},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s", tc.m), func(t *testing.T) {
			fs := m.Find(tc.m)
			if got := len(fs); got != tc.n {
				t.Fatalf("expected %d; got %d", tc.n, got)
			}
		})
	}
}

func TestFiles(t *testing.T) {
	m, err := Open("testdata/mets.xml")
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
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
			fs := m.FindFileGrp(tc.fileGrp)
			if got := fs[0]; got != tc.file {
				t.Fatalf("expected %v; got %v", tc.file, got)
			}
		})
	}
}
