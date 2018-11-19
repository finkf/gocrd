package hocr // import "github.com/finkf/gocrd/xml/hocr"

import (
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"io"
	"strings"
)

// Possible classes for elements
const (
	ClassPage = "ocr_page"
	ClassArea = "ocr_carea"
	ClassLine = "ocr_line"
	ClassWord = "ocrx_word"
)

// ErrNotFound is the error that is returned if
// a attribute of an element could not found.
var ErrNotFound = errors.New("not found")

// Scanner is a low-level scanner for hOCR documents.
type Scanner struct {
	d   *xml.Decoder
	err error
	n   Node
}

// NewScanner creates a new hocr.Scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{d: xml.NewDecoder(r)}
}

// Scan scans the next element in the document.
// It returns true if a new element could be read.
func (s *Scanner) Scan() bool {
	var err error
	var tok xml.Token
	for tok, err = s.d.Token(); tok != nil && err == nil; tok, err = s.d.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			class, _ := findAttr(t.Attr, "class")
			if isValidClass(class) {
				s.n = Element{
					Class: class,
					Node:  t.Copy(),
				}
				return true
			}
		case xml.CharData:
			if s.hasValidNode() {
				s.n = Text(t)
				return true
			}
		}
	}
	return s.handleError(err)
}

// Node returns the last scanned node.
func (s *Scanner) Node() Node {
	return s.n
}

// Err returns the last error.
func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) hasValidNode() bool {
	if e, ok := s.n.(Element); ok {
		return isValidClass(e.Class)
	}
	return false
}

func (s *Scanner) handleError(err error) bool {
	if err == io.EOF {
		return false
	}
	s.err = err
	return s.err == nil
}

func isValidClass(class string) bool {
	switch class {
	case ClassArea, ClassLine, ClassWord, ClassPage:
		return true
	default:
		return false
	}
}

// Node represents hOCR nodes returned by the scanner.
type Node interface{}

// Text is just a typedef for a string.
type Text string

// Element is used to represent text elements in the hOCR document.
type Element struct {
	Class string
	Node  xml.StartElement
}

// BBox returns the bounding box of the element.
// If the element does not have a bounding box,
// the empty boundingbox (0,0)-(0,0) is returned.
func (e Element) BBox() image.Rectangle {
	var r image.Rectangle
	_, err := e.Scanf("title", "bbox", "%d %d %d %d",
		&r.Min.X, &r.Min.Y, &r.Max.X, &r.Max.Y)
	if err != nil {
		return image.Rectangle{}
	}
	return r
}

// Scanf is used to read values of the different element attributes.
// Use like this: e.Scanf("title", "image", "%s", &str)
func (e Element) Scanf(attr, key, format string, args ...interface{}) (int, error) {
	val, found := findAttr(e.Node.Attr, attr)
	if !found {
		return 0, ErrNotFound
	}
	var spos, epos int
	if spos = strings.Index(val, key); spos == -1 {
		return 0, ErrNotFound
	}
	spos += len(key) + 1
	epos = strings.Index(val[spos:], ";")
	if epos == -1 {
		epos = len(val)
	}
	return fmt.Sscanf(val[spos:epos], format, args...)
}

func findAttr(attrs []xml.Attr, name string) (string, bool) {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value, true
		}
	}
	return "", false
}
