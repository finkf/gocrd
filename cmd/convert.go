package cmd

import (
	"encoding/xml"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/finkf/gocrd/xml/hocr"
	"github.com/finkf/gocrd/xml/page"
	"github.com/spf13/cobra"
)

var convertCommand = &cobra.Command{
	Use:   "convert [files...]",
	Short: "converts between different OCR file formats",
	RunE:  convert,
}

var (
	odir string
	from string
	to   string
)

func init() {
	convertCommand.Flags().StringVarP(
		&odir, "odir", "o", "./", "set output directory")
	convertCommand.Flags().StringVarP(
		&from, "from", "f", "", "set source format")
	convertCommand.Flags().StringVarP(
		&to, "to", "t", "", "set destination format")
}

func convert(cmd *cobra.Command, args []string) error {
	// create output directory
	if odir != "./" {
		if err := os.MkdirAll(odir, os.ModePerm); err != nil && !os.IsExist(err) {
			return fmt.Errorf("cannot create ouput directory %s: %v", odir, err)
		}
	}
	// create convert
	c, err := newConverter(from, to)
	if err != nil {
		return err
	}
	// convert input files
	for _, arg := range args {
		if err := c.convert(arg); err != nil {
			return err
		}
	}
	return nil
}

type converter interface {
	convert(string) error
}

func newConverter(from, to string) (converter, error) {
	if from != "OcropyBook" || to != "PageXML" {
		return nil, fmt.Errorf("cannot convert from %s to %s", from, to)
	}
	return &ocropyBookToPageXML{odir: odir}, nil
}

type ocropyBookToPageXML struct {
	odir, bdir string
	scanner    *hocr.Scanner
}

func (c *ocropyBookToPageXML) convert(input string) error {
	is, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("cannot convert %s: %v", input, err)
	}
	defer is.Close()
	c.scanner = hocr.NewScanner(is)
	c.bdir = path.Dir(input)
	if err := c.doConvert(); err != nil {
		return fmt.Errorf("cannot convert %s: %v", input, err)
	}
	return nil
}

func (c *ocropyBookToPageXML) doConvert() error {
	var page *page.PcGts
	var i int
	for c.scanner.Scan() {
		if _, ok := c.scanner.Node().(hocr.Element); !ok { // skip text
			continue
		}
		e := c.scanner.Node().(hocr.Element)
		if e.Class == hocr.ClassPage {
			var err error
			page, err = c.nextPage(page, e)
			if err != nil {
				return err
			}
			i = 0
			continue
		}
		if e.Class == hocr.ClassLine {
			i++
			if err := c.addGTLine(&page.Page, i, e); err != nil {
				return err
			}
		}
	}
	return c.scanner.Err()
}

func (c *ocropyBookToPageXML) addGTLine(p *page.Page, i int, e hocr.Element) error {
	gt, path, err := c.readTrimmedGTLine(p.ImageFilename, i)
	if err != nil {
		return err
	}
	bbox := e.BBox()
	coords := page.Coords{Points: []image.Point{bbox.Min, bbox.Max}}
	line := page.TextLine{
		TextRegionBase: page.TextRegionBase{
			ID:        fmt.Sprintf(p.TextRegion[0].ID+"_l_%d", i),
			Custom:    fmt.Sprintf("gtfile {path:%s}", path),
			Coords:    coords,
			TextEquiv: page.TextEquiv{Unicode: []string{gt}},
		},
	}
	c.addGTWords(&line, gt)
	p.TextRegion[0].TextLine = append(p.TextRegion[0].TextLine, line)
	return nil
}

func (c *ocropyBookToPageXML) addGTWords(l *page.TextLine, gt string) {
	for i, word := range strings.Fields(gt) {
		l.Word = append(l.Word, page.Word{
			TextRegionBase: page.TextRegionBase{
				ID:        fmt.Sprintf(l.ID+"_w_%d", i+1),
				TextEquiv: page.TextEquiv{Unicode: []string{word}},
				Coords:    l.Coords,
			},
		})
	}
}

func (c *ocropyBookToPageXML) readTrimmedGTLine(imgpath string, i int) (gt, file string, err error) {
	imgpath = stripPathExtension(path.Base(imgpath))
	gtfile := path.Join(c.bdir, imgpath, fmt.Sprintf("01%04x.gt.txt", i))
	is, err := os.Open(gtfile)
	if err != nil {
		return "", "", fmt.Errorf("cannot read gt for line %s/%d: %v", imgpath, i, err)
	}
	defer is.Close()
	bgt, err := ioutil.ReadAll(is)
	if err != nil {
		return "", "", fmt.Errorf("cannot read gt for line %s/%d: %v", imgpath, i, err)
	}
	return strings.Trim(string(bgt), " \n\t\r\v"), gtfile, nil

}

func (c *ocropyBookToPageXML) nextPage(old *page.PcGts, e hocr.Element) (*page.PcGts, error) {
	if old != nil {
		if err := c.writePageXML(old); err != nil {
			return nil, err
		}
	}
	var imagepath string
	if _, err := e.Scanf("title", "image", "%s", &imagepath); err != nil {
		return nil, fmt.Errorf("cannot read image path: %v", err)
	}
	bb := e.BBox()
	coords := page.Coords{Points: []image.Point{bb.Min, bb.Max}}
	return &page.PcGts{
		Attributes: page.PcGtsXMLHeader,
		Metadata: page.Metadata{
			"Creator": "GOCRD",
			"Created": time.Now().Format(time.RFC3339),
		},
		Page: page.Page{
			ImageFilename: imagepath,
			ImageHeight:   bb.Max.Y,
			ImageWidth:    bb.Max.X,
			Type:          "content",
			PrintSpace:    page.PrintSpace{Coords: coords},
			TextRegion: []page.TextRegion{
				page.TextRegion{
					TextRegionBase: page.TextRegionBase{
						ID:     "r_1",
						Coords: coords,
					},
					Type: "paragraph",
				},
			},
		},
	}, nil
}

func (c *ocropyBookToPageXML) writePageXML(p *page.PcGts) error {
	opath := path.Join(c.odir, stripPathExtension(path.Base(p.Page.ImageFilename))+".xml")
	xml, err := xml.MarshalIndent(p, "  ", " ")
	if err != nil {
		return fmt.Errorf("cannot write page xml: %v", err)
	}
	if err := ioutil.WriteFile(opath, xml, os.ModePerm); err != nil {
		return fmt.Errorf("cannot write page xml: %v", err)
	}
	return nil
}

func stripPathExtension(path string) string {
	if pos := strings.Index(path, "."); pos != -1 {
		return path[0:pos]
	}
	return path
}
