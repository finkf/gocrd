package page

import (
	"encoding/xml"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"strings"
	"time"
)

// XML namespace, schema instance and location.
const (
	XMLNameSpace      = "http://schema.primaresearch.org/PAGE/gts/pagecontent/2013-07-15"
	XMLSchemaInstance = "http://www.w3.org/2001/XMLSchema-instance"
	XMLSchemaLocation = "http://schema.primaresearch.org/PAGE/gts/pagecontent/2013-07-15" +
		" http://schema.primaresearch.org/PAGE/gts/pagecontent/2013-07-15/pagecontent.xsd"
)

// PcGtsXMLHeader defines the default xml namespace header.
var PcGtsXMLHeader = []xml.Attr{
	xml.Attr{
		Name:  xml.Name{Local: "xmlns"},
		Value: XMLNameSpace,
	},
	xml.Attr{
		Name:  xml.Name{Space: "xmlns", Local: "xsi"},
		Value: XMLSchemaInstance,
	},
	xml.Attr{
		Name:  xml.Name{Space: "xsi", Local: "schemaLocation"},
		Value: XMLSchemaLocation,
	},
}

// PcGts is the top level node of page XML files.
type PcGts struct {
	Attributes []xml.Attr `xml:",attr"`
	Metadata   Metadata   `xml:"Metadata"`
	Page       Page       `xml:"Page"`
}

// Open reads a new page xml file from the given file path.
func Open(path string) (*PcGts, error) {
	is, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer is.Close()
	return Read(is)
}

// Read reads a new page xml file from an input stream.
func Read(r io.Reader) (*PcGts, error) {
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

// UnmarshalXML unmarshals the Metadata of a PcGts structure from xml.
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

// MarshalXML marshals the Metadata of a PcGts structure to xml.
// <Metadata>
// <Creator>OCR-D</Creator>
// ...
// </Metadata>
func (m Metadata) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	selem := xml.StartElement{Name: xml.Name{Local: "Metadata"}}
	if err := e.EncodeToken(selem); err != nil {
		return err
	}
	m["LastChange"] = time.Now().Format(time.RFC3339)
	for k, v := range m {
		name := xml.Name{Local: k}
		if err := e.EncodeToken(xml.StartElement{Name: name}); err != nil {
			return err
		}
		if err := e.EncodeToken(xml.CharData(v)); err != nil {
			return err
		}
		if err := e.EncodeToken(xml.EndElement{Name: name}); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: selem.Name})
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
	ID        string `xml:"id,attr"`
	Custom    string `xml:"custom,attr"`
	Coords    Coords
	TextStyle TextStyle
	TextEquiv TextEquiv
}

// TextRegion is a region of text (paragraph, block, ...)
type TextRegion struct {
	TextRegionBase
	Type     string `xml:"type,attr"`
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
	FontFamaily  string  `xml:"fontFamily,attr,omitempty"`
	Serif        bool    `xml:"serif,attr,omitempty"`
	Monospace    bool    `xml:"monospace,attr,omitempty"`
	FontSize     float32 `xml:"fontSize,attr,omitempty"`
	Kerning      int     `xml:"kerning,attr,omitempty"`
	TextColor    string  `xml:"textColour,attr,omitempty"`
	TextColorRGB int     `xml:"textColourRgb,attr,omitempty"`
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
func (c Coords) BoundingBox() image.Rectangle {
	minx := math.MaxInt64
	maxx := math.MinInt64
	miny := math.MaxInt64
	maxy := math.MinInt64
	for _, p := range c.Points {
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
			var x, y float64
			if _, err := fmt.Sscanf(xy, "%f,%f", &x, &y); err != nil {
				return fmt.Errorf("invalid xy pair %q in %q", xy, attr.Value)
			}
			c.Points = append(c.Points, image.Point{X: int(x), Y: int(y)})
		}
	}
	return d.Skip()
}

// MarshalXML marshals a Coords instance.
// <Coords points="x0,y0 x1,y1 x2,y2,..."/>
func (c *Coords) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	var b strings.Builder
	var pre string
	for _, p := range c.Points {
		b.WriteString(fmt.Sprintf("%s%d,%d", pre, p.X, p.Y))
		pre = " "
	}
	selem := xml.StartElement{
		Name: xml.Name{Local: "Coords"},
		Attr: []xml.Attr{xml.Attr{Name: xml.Name{Local: "points"}, Value: b.String()}},
	}
	if err := e.EncodeToken(selem); err != nil {
		return err
	}
	return e.EncodeToken(xml.EndElement{Name: selem.Name})
}

func ignoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}
