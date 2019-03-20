package cmd

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/finkf/gocrd/xml/page"
	"github.com/spf13/cobra"
)

var convertCommand = &cobra.Command{
	Use:   "convert input-file output-file",
	Short: "converts different file formats to PageXML",
	RunE:  runConvert,
	Args:  cobra.ExactArgs(2),
}

var (
	convertFrom string
)

const (
	formatHOCR = "hOCR"
	// formatOcropy = "Ocropy"
)

func init() {
	convertCommand.Flags().StringVarP(
		&convertFrom, "from", "f", "hOCR", "set source format (hOCR)")
}

func runConvert(cmd *cobra.Command, args []string) (eout error) {
	in, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(args[1])
	if err != nil {
		return err
	}
	defer func() { eout = checkClose(eout, out) }()
	return convert(convertFrom, in, out)
}

func convert(from string, in io.Reader, out io.Writer) error {
	converter, err := newConverter(from)
	if err != nil {
		return err
	}
	return converter.convert(in, out)
}

func newConverter(from string) (converter, error) {
	switch from {
	case formatHOCR:
		return hOCRConverter{}, nil
	// case FormatOcropy:
	// 	return &ocropyBookToPageXML{odir: odir}, nil
	default:
		return nil, fmt.Errorf("cannot convert from %q", from)
	}
}

type converter interface {
	convert(io.Reader, io.Writer) error
}

type hOCRConverter struct{}

func (hOCRConverter) convert(in io.Reader, out io.Writer) error {
	p, err := page.ReadFromHOCR(in)
	if err != nil {
		return err
	}
	e := xml.NewEncoder(out)
	e.Indent("\t", "\t")
	return e.Encode(p)
}

// 	// create output directory
// 	if odir != "./" {
// 		if err := os.MkdirAll(odir, os.ModePerm); err != nil && !os.IsExist(err) {
// 			return fmt.Errorf("cannot create ouput directory %q: %v", odir, err)
// 		}
// 	}
// 	// create convert
// 	c, err := newConverter(from, to)
// 	if err != nil {
// 		return err
// 	}
// 	// convert input files
// 	for _, arg := range args {
// 		if err := c.convert(arg); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// type ocropyBookToPageXML struct {
// 	odir, bdir string
// 	scanner    *hocr.Scanner
// }

// func (c *ocropyBookToPageXML) convert(input string) error {
// 	log.Printf("converting %s", input)
// 	is, err := os.Open(input)
// 	if err != nil {
// 		return fmt.Errorf("cannot convert %q: %v", input, err)
// 	}
// 	defer is.Close()
// 	c.scanner = hocr.NewScanner(is)
// 	c.bdir = path.Dir(input)
// 	if err := c.doConvert(); err != nil {
// 		return fmt.Errorf("cannot convert %q: %v", input, err)
// 	}
// 	return nil
// }

// func (c *ocropyBookToPageXML) doConvert() error {
// 	var page *page.PcGts
// 	var i int
// 	for c.scanner.Scan() {
// 		if _, ok := c.scanner.Node().(hocr.Element); !ok { // skip text
// 			continue
// 		}
// 		e := c.scanner.Node().(hocr.Element)
// 		switch e.Class {
// 		case hocr.ClassPage:
// 			var err error
// 			page, err = c.nextPage(page, e)
// 			if err != nil {
// 				return err
// 			}
// 			i = 0
// 		case hocr.ClassLine:
// 			i++
// 			if err := c.addGTLine(&page.Page, i, e); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return c.scanner.Err()
// }

// func (c *ocropyBookToPageXML) addGTLine(p *page.Page, i int, e hocr.Element) error {
// 	gt, path, err := c.readTrimmedGTLine(p.ImageFilename, i)
// 	if err != nil {
// 		return err
// 	}
// 	// append lines to text region
// 	if len(p.TextRegion[0].TextEquiv.Unicode) == 0 {
// 		p.TextRegion[0].TextEquiv.Unicode = append(
// 			p.TextRegion[0].TextEquiv.Unicode, gt)
// 	} else {
// 		p.TextRegion[0].TextEquiv.Unicode[0] += "\n" + gt
// 	}
// 	// add line
// 	bbox := e.BBox()
// 	coords := page.Coords{Points: []image.Point{bbox.Min, bbox.Max}}
// 	line := page.TextLine{
// 		TextRegionBase: page.TextRegionBase{
// 			ID:        fmt.Sprintf(p.TextRegion[0].ID+"_l_%d", i),
// 			Custom:    fmt.Sprintf("gtfile {path:%s}", path),
// 			Coords:    coords,
// 			TextEquiv: page.TextEquiv{Unicode: []string{gt}},
// 		},
// 	}
// 	// add words
// 	c.addGTWords(&line, gt)
// 	p.TextRegion[0].TextLine = append(p.TextRegion[0].TextLine, line)
// 	return nil
// }

// func (c *ocropyBookToPageXML) addGTWords(l *page.TextLine, gt string) {
// 	for i, word := range strings.Fields(gt) {
// 		l.Word = append(l.Word, page.Word{
// 			TextRegionBase: page.TextRegionBase{
// 				ID:        fmt.Sprintf(l.ID+"_w_%d", i+1),
// 				TextEquiv: page.TextEquiv{Unicode: []string{word}},
// 				Coords:    l.Coords,
// 			},
// 		})
// 	}
// }

// func (c *ocropyBookToPageXML) readTrimmedGTLine(imgpath string, i int) (gt, file string, err error) {
// 	imgpath = stripPathExtension(path.Base(imgpath))
// 	gtfile := path.Join(c.bdir, imgpath, fmt.Sprintf("01%04x.gt.txt", i))
// 	is, err := os.Open(gtfile)
// 	if err != nil {
// 		return "", "", fmt.Errorf("cannot read gt for line %s/%d: %v", imgpath, i, err)
// 	}
// 	defer is.Close()
// 	bgt, err := ioutil.ReadAll(is)
// 	if err != nil {
// 		return "", "", fmt.Errorf("cannot read gt for line %s/%d: %v", imgpath, i, err)
// 	}
// 	return strings.Trim(string(bgt), " \n\t\r\v"), path.Join(imgpath, path.Base(gtfile)), nil
// }

// func (c *ocropyBookToPageXML) nextPage(old *page.PcGts, e hocr.Element) (*page.PcGts, error) {
// 	if old != nil {
// 		if err := c.writePageXML(old); err != nil {
// 			return nil, err
// 		}
// 	}
// 	var imagepath string
// 	if _, err := e.Scanf("title", "image", "%s", &imagepath); err != nil {
// 		if _, err := e.Scanf("title", "file", "%s", &imagepath); err != nil {
// 			return nil, fmt.Errorf("cannot read image path: %v", err)
// 		}
// 	}
// 	bb := e.BBox()
// 	coords := page.Coords{Points: []image.Point{bb.Min, bb.Max}}
// 	return &page.PcGts{
// 		Attributes: page.PcGtsXMLHeader,
// 		Metadata: page.Metadata{
// 			"Creator": "GOCRD",
// 			"Created": time.Now().Format(time.RFC3339),
// 		},
// 		Page: page.Page{
// 			ImageFilename: imagepath,
// 			ImageHeight:   bb.Max.Y,
// 			ImageWidth:    bb.Max.X,
// 			Type:          "content",
// 			PrintSpace:    page.PrintSpace{Coords: coords},
// 			TextRegion: []page.TextRegion{
// 				page.TextRegion{
// 					TextRegionBase: page.TextRegionBase{
// 						ID:     "r_1",
// 						Coords: coords,
// 					},
// 					Type: "paragraph",
// 				},
// 			},
// 		},
// 	}, nil
// }

// func (c *ocropyBookToPageXML) writePageXML(p *page.PcGts) error {
// 	opath := path.Join(c.odir, stripPathExtension(path.Base(p.Page.ImageFilename))+".xml")
// 	out, err := os.Open(opath)
// 	if err != nil {
// 		return fmt.Errorf("cannot write page xml: %v", err)
// 	}
// 	defer out.Close()
// 	e := xml.NewEncoder(out)
// 	e.Indent("\t", "\t")
// 	if err := e.Encode(p); err != nil {
// 		return fmt.Errorf("cannot write page xml: %v", err)
// 	}
// 	return nil
// }

// func stripPathExtension(path string) string {
// 	if pos := strings.Index(path, "."); pos != -1 {
// 		return path[0:pos]
// 	}
// 	return path
// }
