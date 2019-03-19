package cmd

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"sort"
	"strings"

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
	catCommand.Flags().StringVarP(
		&synpageOutput, "out", "o", synpageOutput, "base name for output files")
	catCommand.Flags().BoolVarP(
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
	log.Printf("rect: %s", rect)
	return nil
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
	log.Printf("new image %s", rec)
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
	draw.Draw(rgba, old.Bounds(), old, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, dest, new, image.Point{0, 0}, draw.Src)
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
	sp.page.Metadata = make(page.Metadata)
	out, err := os.Create(synpageOutput + ".xml")
	if err != nil {
		return fmt.Errorf("cannot write pageXML: %v", err)
	}
	defer func() { eout = checkClose(eout, out) }()
	if err := xml.NewEncoder(out).Encode(sp.page); err != nil {
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
