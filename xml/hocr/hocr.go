package hocr // import "github.com/finkf/gocrd/xml/hocr"

import (
	"io"
	"os"
)

// HOCR mimics the structure of a HOCR file
type HOCR struct {
}

// Read reads HOCR from a reader.
func Read(r io.Reader) (*HOCR, error) {
	return nil, nil
}

// Open opens a HOCR file and reads it.
func Open(p string) (*HOCR, error) {
	is, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer is.Close()
	return Read(is)
}
