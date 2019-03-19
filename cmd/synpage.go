package cmd

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/finkf/gocrd/boundingbox"
	"github.com/finkf/gocrd/xml/page"
	"github.com/spf13/cobra"
)

var synpageCommand = &cobra.Command{
	Use:   "synpage [flags] [line-files...]",
	Short: "Build synthetic page files from ocropy training-lines",
	Long: `Build synthetic page files from Ocropy training-lines.
Convert Orcorpy line files and line images to a PageXML file
with an according image files.`,
	RunE: runSynpage,
}

var (
	synpageOutput = "out"
	synpageSort   = false
)

func init() {
	synpageCommand.Flags().StringVarP(
		&synpageOutput, "out", "o", synpageOutput, "base name for output files")
	synpageCommand.Flags().BoolVarP(
		&synpageSort, "sort", "s", synpageSort, "sort input files")
}

func runSynpage(cmd *cobra.Command, args []string) error {
	return synpage(args, os.Stdout)
}

func synpage(args []string, stdout io.Writer) error {
	if synpageSort {
		sort.Strings(args)
	}
	var sp synpageS
	for _, arg := range args {
		if err := sp.addLine(arg); err != nil {
			return err
		}
	}
	return sp.write()
}

type synpageS struct {
	page page.PcGts
	img  image.Image
}

func (sp *synpageS) addLine(path string) error {
	img, err := openLineImage(path)
	if err != nil {
		return fmt.Errorf("cannot open line image: %v", err)
	}
	rect := sp.appendImage(img)
	text, err := sp.readLine(path)
	if err != nil {
		return fmt.Errorf("cannot read text line: %v", err)
	}
	sp.appendLineRegion(text, rect)
	return nil
}

func (sp *synpageS) appendLineRegion(text string, rect image.Rectangle) {
	if len(sp.page.Page.TextRegion) == 0 {
		sp.page.Page.TextRegion = append(sp.page.Page.TextRegion, page.TextRegion{
			TextRegionBase: page.TextRegionBase{ID: "r_1"},
		})
	}
	lineID := len(sp.page.Page.TextRegion[0].TextLine) + 1
	line := page.TextLine{
		TextRegionBase: page.TextRegionBase{
			ID:        fmt.Sprintf("r_1_l_%d", lineID),
			Coords:    page.Coords{Points: []image.Point{rect.Min, rect.Max}},
			TextEquiv: page.TextEquiv{Unicode: []string{text}},
		},
		BaseLine: page.Coords{Points: baseline(rect)},
	}
	var cut int
	y0 := rect.Min.Y
	y1 := rect.Max.Y
	for i, split := range boundingbox.SplitTokens(rect, text) {
		word := page.Word{
			TextRegionBase: page.TextRegionBase{
				ID:        fmt.Sprintf("%s_w_%d", line.ID, i+1),
				Coords:    page.Coords{Points: []image.Point{{cut, y0}, {split.Cut, y1}}},
				TextEquiv: page.TextEquiv{Unicode: []string{split.Str}},
			},
		}
		cut = split.Cut
		line.Word = append(line.Word, word)
	}
	sp.page.Page.TextRegion[0].TextLine = append(sp.page.Page.TextRegion[0].TextLine, line)
}

func (sp *synpageS) readLine(path string) (string, error) {
	is, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer is.Close()
	s := bufio.NewScanner(is)
	if s.Scan() {
		return s.Text(), nil
	}
	return "", s.Err()
}

func (sp *synpageS) appendImage(new image.Image) image.Rectangle {
	if sp.img == nil {
		sp.img = new
		return sp.img.Bounds()
	}
	old := sp.img
	rec := old.Bounds()
	if new.Bounds().Max.X > old.Bounds().Max.X {
		rec.Max.X = new.Bounds().Max.X
	}
	rec.Max.Y += new.Bounds().Max.Y
	rgba := image.NewRGBA(rec)
	dest := image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: old.Bounds().Max.Y,
		},
		Max: image.Point{
			X: new.Bounds().Max.X,
			Y: old.Bounds().Max.Y + new.Bounds().Max.Y,
		},
	}
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
	draw.Draw(rgba, old.Bounds(), old, image.ZP, draw.Src)
	draw.Draw(rgba, dest, new, image.ZP, draw.Src)
	sp.img = rgba
	return dest
}

func (sp *synpageS) write() error {
	if err := sp.writePageXML(); err != nil {
		return err
	}
	return sp.writeImage()
}

func (sp *synpageS) writePageXML() (eout error) {
	sp.page.Metadata = page.Metadata{
		"Creator":    "GOCRD-synpage",
		"Created":    time.Now().Format(time.RFC3339),
		"LastChange": time.Now().Format(time.RFC3339),
	}
	if len(sp.page.Page.TextRegion) > 0 {
		sp.page.Page.TextRegion[0].Coords = page.Coords{
			Points: []image.Point{sp.img.Bounds().Min, sp.img.Bounds().Max},
		}
		sp.page.Page.TextRegion[0].TextEquiv = page.TextEquiv{
			Unicode: []string{""},
		}
		for i, line := range sp.page.Page.TextRegion[0].TextLine {
			if i != 0 {
				sp.page.Page.TextRegion[0].TextEquiv.Unicode[0] += "\n"
			}
			sp.page.Page.TextRegion[0].TextEquiv.Unicode[0] += line.TextEquiv.Unicode[0]
		}
	}
	xpath := synpageOutput + ".xml"
	ipath := synpageOutput + ".png"
	sp.page.Page.ImageFilename = ipath
	sp.page.Page.ImageWidth = sp.img.Bounds().Dx()
	sp.page.Page.ImageHeight = sp.img.Bounds().Dy()
	out, err := os.Create(xpath)
	if err != nil {
		return fmt.Errorf("cannot write pageXML: %v", err)
	}
	defer func() { eout = checkClose(eout, out) }()
	e := xml.NewEncoder(out)
	e.Indent("\t", "\t")
	if err := e.Encode(sp.page); err != nil {
		return fmt.Errorf("cannot write pageXML: %v", err)
	}
	return nil
}

func (sp *synpageS) writeImage() (eout error) {
	// write image
	out, err := os.Create(synpageOutput + ".png")
	if err != nil {
		return fmt.Errorf("cannot write image: %v", err)
	}
	defer func() { eout = checkClose(eout, out) }()
	if err := png.Encode(out, sp.img); err != nil {
		return fmt.Errorf("cannot write image: %v", err)
	}
	return nil
}

func checkClose(e1 error, c io.Closer) error {
	e2 := c.Close()
	if e1 != nil {
		return e1
	}
	return e2
}

func openLineImage(path string) (image.Image, error) {
	imgpath := path
	if pos := strings.Index(path, "."); pos != -1 {
		imgpath = path[0:pos]
	}
	for _, ext := range []string{".bin.png", ".dew.png", ".png"} {
		if _, err := os.Stat(imgpath + ext); err == nil {
			is, err := os.Open(imgpath + ext)
			if err != nil {
				return nil, fmt.Errorf("cannot open image file: %v", err)
			}
			defer is.Close()
			return png.Decode(is)
		}
	}
	return nil, fmt.Errorf("cannot find image file for %q", path)
}

func baseline(rect image.Rectangle) []image.Point {
	y0 := rect.Min.Y + (rect.Dy() / 2)
	return []image.Point{{rect.Min.X, y0}, {rect.Max.X, y0}}
}
