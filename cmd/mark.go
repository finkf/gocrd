package cmd

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"

	"github.com/finkf/gocrd/xml/page"
	"github.com/spf13/cobra"
	"golang.org/x/image/tiff"
)

var markCommand = &cobra.Command{
	Use:   "mark pagexml-file image-file",
	Short: "mark text regions in an image",
	RunE:  runMark,
	Args:  cobra.ExactArgs(3),
}
var markLevel = "line"

func init() {
	markCommand.Flags().StringVarP(
		&markLevel, "level", "l", markLevel,
		"set level of regions to mark (region, line, word, glyph)")
}

func runMark(cmd *cobra.Command, args []string) error {
	return mark(args[0], args[1], args[2])
}

func mark(pagexml, image, outpath string) error {
	p, err := page.Open(pagexml)
	if err != nil {
		return err
	}
	is, err := os.Open(image)
	if err != nil {
		return err
	}
	defer is.Close()
	img, err := tiff.Decode(is)
	if err != nil {
		return err
	}
	if err = markRegions(p, img.(draw.Image)); err != nil {
		return err
	}
	out, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer out.Close()
	return tiff.Encode(out, img, &tiff.Options{})
}

func markRegions(p *page.PcGts, img draw.Image) error {
	mr := strings.ToLower(markLevel) == "region"
	ml := strings.ToLower(markLevel) == "line"
	mw := strings.ToLower(markLevel) == "word"
	mg := strings.ToLower(markLevel) == "glyph"
	for _, r := range p.Page.TextRegion {
		if mr {
			drawRectangle(r.Coords.BoundingBox(), img)
			continue
		}
		for _, l := range r.TextLine {
			if ml {
				drawRectangle(l.Coords.BoundingBox(), img)
			}
			for _, w := range l.Word {
				if mw {
					drawRectangle(w.Coords.BoundingBox(), img)
					continue
				}
				for _, g := range w.Glyph {
					if mg {
						drawRectangle(g.Coords.BoundingBox(), img)
					}
				}
			}
		}
	}
	return nil
}

var col = color.RGBA{255, 0, 0, 255}

func drawRectangle(r image.Rectangle, img draw.Image) {
	rect(img, r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
}

func hline(img draw.Image, x1, y, x2 int) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}

func vline(img draw.Image, x, y1, y2 int) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}

func rect(img draw.Image, x1, y1, x2, y2 int) {
	hline(img, x1, y1, x2)
	hline(img, x1, y2, x2)
	vline(img, x1, y1, y2)
	vline(img, x2, y1, y2)
}
