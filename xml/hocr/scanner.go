package hocr // import "github.com/finkf/gocrd/xml/hocr"

import (
	"encoding/xml"
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

// Scanner is a low-level scanner for hOCR documents.
type Scanner struct {
	d     *xml.Decoder
	err   error
	node  Node
	stack stack
}

// NewScanner creates a new hocr.Scanner
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{d: xml.NewDecoder(r)}
}

// Scan scans the next element in the document.  It returns true if a
// new element was scanned and false if an error occured or if there
// is no more nodes to be scanned.
func (s *Scanner) Scan() bool {
	var err error
	var tok xml.Token
	for tok, err = s.d.Token(); tok != nil && err == nil; tok, err = s.d.Token() {
		switch t := tok.(type) {
		case xml.StartElement:
			cont, ret := s.handleStartElement(t)
			if !cont {
				return ret
			}
		case xml.CharData:
			if s.stack.match("html", "head", "title") {
				s.node = Title(t)
				return true
			}
			if s.hasValidNode() {
				str := strings.Trim(string(t), "\n\r\t\v ")
				if str != "" {
					s.node = Text(str)
					return true
				}
				continue
			}
		case xml.EndElement:
			s.stack = s.stack.pop()
		}
	}
	return s.handleError(err)
}

// Node returns the last scanned node.
func (s *Scanner) Node() Node {
	return s.node
}

// Err returns the last error.
func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) handleStartElement(t xml.StartElement) (cont, ret bool) {
	s.stack = s.stack.push(t.Name.Local)
	// /html/head/meta tag
	if s.stack.match("html", "head", "meta") {
		node := s.parseMeta(t)
		if node != nil {
			s.node = node
			return false, true
		}
		return true, false
	}
	// an element with class="..."
	class, _ := findAttr(t.Attr, "class")
	if isValidClass(class) {
		s.node = Element{
			Class: class,
			Node:  t.Copy(),
		}
		return false, true
	}
	// */p element
	if s.stack.match("p") {
		s.node = Paragraph{}
		return false, true
	}
	return true, false
}

func (s *Scanner) parseMeta(elem xml.StartElement) Node {
	name, ok := findAttr(elem.Attr, "name")
	if !ok {
		return nil
	}
	content, _ := findAttr(elem.Attr, "content")
	return Meta{Name: name, Content: content}
}

func (s *Scanner) hasValidNode() bool {
	if e, ok := s.node.(Element); ok {
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

// Node represents hOCR nodes returned by the scanner.  A Node is
// either of type Text, Element, Paragraph, Title or Meta.
type Node interface{}

// Text is used to represent (non empty) char data nodes.
type Text string

// Title represents the char data nodes of /html/head/title elements.
type Title string

// Paragraph represents naked <p> tags without any class informations.
type Paragraph struct{}

// Meta represents /html/head/meta tags.
type Meta struct {
	Name, Content string
}

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
		return 0, fmt.Errorf("cannot find attribute: %s", attr)
	}
	var spos, epos int
	if spos = strings.Index(val, key); spos == -1 {
		return 0, fmt.Errorf("cannot find key: %s", key)
	}
	spos += len(key) + 1
	epos = strings.Index(val[spos:], ";")
	if epos == -1 {
		epos = len(val)
	} else {
		epos += spos
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
