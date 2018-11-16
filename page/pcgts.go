package page

import (
	"encoding/xml"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"strings"
)

//<PcGts xmlns="http://schema.primaresearch.org/PAGE/gts/pagecontent/2013-07-15" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://schema.primaresearch.org/PAGE/gts/pagecontent/2013-07-15 http://schema.primaresearch.org/PAGE/gts/pagecontent/2013-07-15/pagecontent.xsd">

// PcGts is the top level node of page XML files.
type PcGts struct {
	Attributes []xml.Attr
	Metadata   Metadata `xml:"Metadata"`
	Page       Page     `xml:"page"`
}

// OpenPcGts reads a new page xml file from the given file path.
func OpenPcGts(path string) (*PcGts, error) {
	is, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer is.Close()
	return ReadPcGts(is)
}

// ReadPcGts reads a new page xml file from an input stream.
func ReadPcGts(r io.Reader) (*PcGts, error) {
	var p PcGts
	p.Metadata = make(Metadata)
	if err := xml.NewDecoder(r).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// UnmarshalXML unmarshals the top-level PcGts node of page xml files.
func (p *PcGts) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Attributes = start.Attr
	var err error
	var t xml.Token
	for t, err = d.Token(); t != nil && err == nil; t, err = d.Token() {
		switch tt := t.(type) {
		case xml.StartElement:
			switch tt.Name.Local {
			case "Metadata":
				d.DecodeElement(&p.Metadata, &tt)
			case "Page":
				if err = d.DecodeElement(&p.Page, &tt); err != nil {
					return err
				}
			}
		}
	}
	return ignoreEOF(err)
}

// Metadata defines
type Metadata map[string]string

// UnmarshalXML unmarshals the Metadata of a PcGts structure.
func (m Metadata) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var node string
	var err error
	var t xml.Token
	for t, err = d.Token(); t != nil && err == nil; t, err = d.Token() {
		switch tt := t.(type) {
		case xml.StartElement:
			node = tt.Name.Local
		case xml.CharData:
			if node == "" {
				continue
			}
			m[node] = string(tt)
		case xml.EndElement:
			node = ""
		}
	}
	return ignoreEOF(err)
}

// Page is a page in a PcGts structure.
type Page struct {
	ImageFilename string `xml:"imageFilename,attr"`
	ImageHeight   int    `xml:"imageHeight,attr"`
	ImageWidth    int    `xml:"imageWidth,attr"`
	Type          string `xml:"type,attr"`
	PrintSpace    PrintSpace
	ReadingOrder  ReadingOrder
	TextRegion    []TextRegion
}

// PrintSpace defines the print space of a page.
type PrintSpace struct {
	Coords Coords
}

// ReadingOrder is a collection of ordered groups.
type ReadingOrder struct {
	OrderedGroup []OrderedGroup
}

// OrderedGroup is a collection of regions.
type OrderedGroup struct {
	ID               string `xml:"id,attr"`
	Caption          string `xml:"caption,attr"`
	RegionRefIndexed []RegionRefIndexed
}

// RegionRefIndexed is a index region.
type RegionRefIndexed struct {
	Index     int    `xml:"index,attr"`
	RegionRef string `xml:"regionRef,attr"`
}

// TextRegionBase defines the base data structure for
// all text regions (TextRegion, Line, Word, Glyph) in a page XML document.
type TextRegionBase struct {
	Type      string `xml:"type,attr"`
	ID        string `xml:"id,attr"`
	Custom    string `xml:"custom,attr"`
	Coords    Coords
	TextStyle TextStyle
	TextEquiv TextEquiv
}

// TextRegion is a region of text (paragraph, block, ...)
type TextRegion struct {
	TextRegionBase
	TextLine []TextLine
}

// TextLine is a line of text in a text region.
type TextLine struct {
	TextRegionBase
	BaseLine Coords `xml:"Baseline"`
	Word     []Word
}

// Word is a token in a line.
type Word struct {
	TextRegionBase
	Glyph []Glyph
}

// Glyph is a single character in a word.
type Glyph struct {
	TextRegionBase
}

// TextStyle specifies font information of any text region.
type TextStyle struct {
	FontFamaily  string  `xml:"fontFamily,attr"`
	Serif        bool    `xml:"serif,attr"`
	Monospace    bool    `xml:"monospace,attr"`
	FontSize     float32 `xml:"fontSize,attr"`
	Kerning      int     `xml:"kerning,attr"`
	TextColor    string  `xml:"textColour,attr"`
	TextColorRGB int     `xml:"textColourRgb,attr"`
	/* ... */
}

// TextEquiv defines the text string of text regions.
type TextEquiv struct {
	PlainText []string
	Unicode   []string
}

// Coords are rectangles of points.
type Coords struct {
	Points []image.Point `xml:"points,attr"`
}

// BoundingBox returns the bounding box of the polygon.
func (p Coords) BoundingBox() image.Rectangle {
	minx := math.MaxInt64
	maxx := math.MinInt64
	miny := math.MaxInt64
	maxy := math.MinInt64
	for _, p := range p.Points {
		if p.X < minx {
			minx = p.X
		}
		if p.Y < miny {
			miny = p.Y
		}
		if p.X > maxx {
			maxx = p.X
		}
		if p.Y > maxy {
			maxy = p.Y
		}
	}
	return image.Rect(int(minx), int(miny), int(maxx), int(maxy))
}

// UnmarshalXML unmarshals a Coords instance.
func (c *Coords) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		if attr.Name.Local != "points" {
			continue
		}
		for _, xy := range strings.Fields(attr.Value) {
			var x, y int
			if _, err := fmt.Sscanf(xy, "%d,%d", &x, &y); err != nil {
				return fmt.Errorf("invalid xy pair %q in %q", xy, attr.Value)
			}
			c.Points = append(c.Points, image.Point{X: x, Y: y})
		}
	}
	return d.Skip()
}

func ignoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}
