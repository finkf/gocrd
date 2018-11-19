package hocr

import (
	"fmt"
	"image"
	"os"
	"testing"
)

func withOpenHOCRScanner(path string, f func(*Scanner)) {
	is, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer is.Close()
	s := NewScanner(is)
	f(s)
}

func want(fmt string, got, want interface{}, t *testing.T) {
	t.Helper()
	if got != want {
		t.Fatalf("expected "+fmt+"; got "+fmt, want, got)
	}
}

func TestScannerBBox(t *testing.T) {
	tests := []struct {
		class string
		want  image.Rectangle
	}{
		{ClassPage, image.Rect(0, 0, 2076, 2952)},
		{ClassLine, image.Rect(384, 122, 1712, 206)},
		{ClassLine, image.Rect(1383, 2406, 1703, 2491)},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s", tc.want), func(t *testing.T) {
			withOpenHOCRScanner("testdata/test.html", func(s *Scanner) {
				for s.Scan() {
					if e, ok := s.Node().(Element); !ok || e.Class != tc.class {
						continue
					}
					bb := s.Node().(Element).BBox()
					if bb == tc.want {
						return
					}
				}
				if s.Err() != nil {
					t.Fatalf("got error: %s", s.Err())
				}
				t.Fatalf("cannot find: %v", tc)
			})
		})
	}
}
