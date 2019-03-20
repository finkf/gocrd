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

func TestScanTitle(t *testing.T) {
	withOpenHOCRScanner("testdata/test.html", func(s *Scanner) {
		want := "OCR Results"
		for s.Scan() {
			title, ok := s.Node().(Title)
			if !ok {
				continue
			}
			if string(title) != want {
				t.Fatalf("expected %s; got %s", want, title)
			}
			break
		}
		if s.Err() != nil {
			t.Fatalf("got error: %v", s.Err())
		}
	})
}

func TestScanLines(t *testing.T) {
	tests := []struct {
		want string
		i    int
	}{
		{"muͤts / welches ich nebend anderem gunſt", 1},
		{"vñ freiindtlicher zůͤredte / ſo ſy mir vnerkan⸗", 2},
		{"ſchenckten", 29},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s", tc.want), func(t *testing.T) {
			withOpenHOCRScanner("testdata/test.html", func(s *Scanner) {
				var i int
				for s.Scan() {
					if e, ok := s.Node().(Element); !ok || e.Class != ClassLine {
						continue
					}
					i++
					if i == tc.i {
						if !s.Scan() {
							t.Fatalf("cannot read char data for line %d", tc.i)
						}
						if got := string(s.Node().(Text)); got != tc.want {
							t.Fatalf("expected %s; got %s", tc.want, got)
						}
						break
					}
				}
				if s.Err() != nil {
					t.Fatalf("got error: %v", s.Err())
				}
			})
		})
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
					bb := s.Node().(Element).BoundingBox()
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
