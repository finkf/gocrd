package page

import (
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"github.com/finkf/gocrd/xml/hocr"
)

// OpenFromHOCR reads a hOCR file.  Returns the hOCR content as
// PageXML structure.  This method assumes one page per hOCR document.
func OpenFromHOCR(file string) (*PcGts, error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("cannot open hocr file %q: %v", file, err)
	}
	defer in.Close()
	p, err := ReadFromHOCR(in)
	if err != nil {
		p.Metadata["Origin"] = path.Base(file)
	}
	return p, err
}

// ReadFromHOCR parses a hOCR file.  Returns the hOCR content as
// PageXML structure.  This method assumes one page per hOCR document.
func ReadFromHOCR(in io.Reader) (*PcGts, error) {
	p := &PcGts{
		Attributes: PcGtsXMLHeader,
		Metadata: Metadata{
			"Created":     time.Now().Format(time.RFC3339),
			"LastChanged": time.Now().Format(time.RFC3339),
			"Creator":     "GOCRD-CONVERTED-FROM-HOCR",
		},
	}
	s := hocr.NewScanner(in)
	var prevElem hocr.Element
	for s.Scan() {
		switch t := s.Node().(type) {
		case hocr.Text:
			addHOCRText(p, prevElem, string(t))
		case hocr.Element:
			prevElem = t
			addHOCRRegion(p, t)
		case hocr.Title:
			p.Metadata["Title"] = string(t)
		case hocr.Meta:
			p.Metadata[t.Name] = t.Content
		}
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	hOCRUpateTextEquivs(p)
	return p, nil
}

func addHOCRRegion(p *PcGts, elem hocr.Element) {
	switch elem.Class {
	case hocr.ClassPage:
		handleHOCRPage(p, elem)
	case hocr.ClassArea:
		handleHOCRArea(p, elem)
	case hocr.ClassLine:
		handleHOCRLine(p, elem)
	case hocr.ClassWord:
		handleHOCRWord(p, elem)
	}
}

func handleHOCRPage(p *PcGts, elem hocr.Element) {
	// try to get image file name
	for _, name := range []string{"file", "image", "imagefile"} {
		var imagefile string
		if elem.Scanf("title", name, "%s", &imagefile) {
			p.Page.ImageFilename = imagefile
			break
		}
	}
	// get bounding box
	rect := elem.BoundingBox()
	p.Page.ImageWidth = rect.Dx()
	p.Page.ImageHeight = rect.Dy()
}

func handleHOCRArea(p *PcGts, elem hocr.Element) {
	class, _ := elem.Attribute("class")
	region := TextRegion{
		TextRegionBase: newTextRegionBase(elem),
		Type:           class,
	}
	p.Page.TextRegion = append(p.Page.TextRegion, region)
}

func handleHOCRLine(p *PcGts, elem hocr.Element) {
	if len(p.Page.TextRegion) == 0 {
		p.Page.TextRegion = append(p.Page.TextRegion, TextRegion{})
	}
	i := len(p.Page.TextRegion) - 1
	line := TextLine{
		TextRegionBase: newTextRegionBase(elem),
	}
	var arctan float64
	var yaxis int
	if elem.Scanf("title", "baseline", "%f %d", &arctan, yaxis) {
		line.BaseLine.Points = hOCRBaseLine(line.Coords.BoundingBox(), arctan, yaxis)
	}
	p.Page.TextRegion[i].TextLine = append(p.Page.TextRegion[i].TextLine, line)
}

func handleHOCRWord(p *PcGts, elem hocr.Element) {
	if len(p.Page.TextRegion) == 0 {
		p.Page.TextRegion = append(p.Page.TextRegion, TextRegion{})
	}
	i := len(p.Page.TextRegion) - 1
	if len(p.Page.TextRegion[i].TextLine) == 0 {
		p.Page.TextRegion[i].TextLine = append(p.Page.TextRegion[i].TextLine, TextLine{})
	}
	j := len(p.Page.TextRegion[i].TextLine) - 1
	word := Word{
		TextRegionBase: newTextRegionBase(elem),
	}
	p.Page.TextRegion[i].TextLine[j].Word = append(p.Page.TextRegion[i].TextLine[j].Word, word)
}

func addHOCRText(p *PcGts, elem hocr.Element, text string) {
	te := TextEquiv{Unicode: []string{text}}
	var conf float64
	if elem.Scanf("title", "x_wconf", "%f", &conf) {
		te.Conf = 1.0 / conf
	}
	// skip if no regions of text where encountered (no lines, no words)
	if len(p.Page.TextRegion) == 0 {
		return
	}
	// add text (if the previous element was a line or word)
	i := len(p.Page.TextRegion) - 1
	switch elem.Class {
	case hocr.ClassLine:
		j := len(p.Page.TextRegion[i].TextLine) - 1
		p.Page.TextRegion[i].TextLine[j].TextEquiv = te
	case hocr.ClassWord:
		j := len(p.Page.TextRegion[i].TextLine) - 1
		k := len(p.Page.TextRegion[i].TextLine[j].Word) - 1
		p.Page.TextRegion[i].TextLine[j].Word[k].TextEquiv = te
	}
}

func hOCRUpateTextEquivs(p *PcGts) {
	for i := range p.Page.TextRegion {
		lines := make([]string, len(p.Page.TextRegion[i].TextLine))
		for j := range p.Page.TextRegion[i].TextLine {
			if len(p.Page.TextRegion[i].TextLine[j].TextEquiv.Unicode) == 0 {
				words := make([]string, len(p.Page.TextRegion[i].TextLine[j].Word))
				for k, word := range p.Page.TextRegion[i].TextLine[j].Word {
					if len(word.TextEquiv.Unicode) > 0 {
						words[k] = word.TextEquiv.Unicode[0]
					}
				}
				p.Page.TextRegion[i].TextLine[j].TextEquiv.Unicode =
					[]string{strings.Join(words, " ")}
			}
			lines[j] = p.Page.TextRegion[i].TextLine[j].TextEquiv.Unicode[0]
		}
		if len(p.Page.TextRegion[i].TextEquiv.Unicode) == 0 {
			p.Page.TextRegion[i].TextEquiv.Unicode =
				[]string{strings.Join(lines, "\n")}
		}
	}
}

func newTextRegionBase(elem hocr.Element) TextRegionBase {
	rect := elem.BoundingBox()
	id, _ := elem.Attribute("id")
	custom, _ := elem.Attribute("title")
	return TextRegionBase{
		ID:     id,
		Coords: Coords{Points: []image.Point{rect.Min, rect.Max}},
		Custom: custom,
	}
}

func hOCRBaseLine(r image.Rectangle, arctan float64, yaxis int) []image.Point {
	x0 := r.Min.X
	y0 := r.Min.Y + yaxis
	x1 := r.Max.X
	y1 := int(((math.Pi * float64(180*arctan)) / float64(x1-x0)) + float64(y0))
	return []image.Point{{x0, y0}, {x1, y1}}
}
