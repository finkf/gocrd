package cmd

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"github.com/finkf/gocrd/xml/hocr"
	"github.com/finkf/gocrd/xml/page"
	"github.com/finkf/lev"
	"github.com/spf13/cobra"
)

var replaceCommand = &cobra.Command{
	Use:   "replace",
	Short: "replace text in an OCR file",
	RunE:  runReplace,
}

func runReplace(cmd *cobra.Command, args []string) error {
	gtlines, err := readGTLines(args[0])
	if err != nil {
		return err
	}
	ocrlines, err := readOCRLines(args[1])
	if err != nil {
		return err
	}
	for i := 0; i < len(ocrlines); i++ {
		gtline, _ := bestFit(gtlines, ocrlines[i].str)
		// fmt.Printf("%s\t%s\t%d\t%d\t%d\t%d\t%d\n",
		// 	gtline,
		// 	ocrlines[i].str,
		// 	ocrlines[i].bb.Min.X,
		// 	ocrlines[i].bb.Min.Y,
		// 	ocrlines[i].bb.Max.X,
		// 	ocrlines[i].bb.Max.Y,
		// 	d)
		ocrlines[i].str = gtline
	}
	printSpace := imageRectangle(ocrlines)
	var r page.TextRegion
	r.TextRegionBase.ID = "r_1"
	r.TextRegionBase.Coords.Points = printSpace
	r.TextRegionBase.TextEquiv.Unicode = []string{textRegionString(ocrlines)}
	r.Type = "content"
	for i, ocrline := range ocrlines {
		var l page.TextLine
		l.TextRegionBase.ID = fmt.Sprintf("r_1_l_%d", i+1)
		l.TextRegionBase.Coords.Points = []image.Point{ocrline.bb.Min, ocrline.bb.Max}
		l.TextRegionBase.TextEquiv.Unicode = []string{ocrline.str}
		for j, word := range strings.Split(ocrline.str, " ") {
			var w page.Word
			w.TextRegionBase.ID = fmt.Sprintf("r_1_l_%d_w_%d", i+1, j+1)
			w.TextRegionBase.Coords.Points = []image.Point{ocrline.bb.Min, ocrline.bb.Max}
			w.TextRegionBase.TextEquiv.Unicode = []string{word}
			l.Word = append(l.Word, w)
		}
		r.TextLine = append(r.TextLine, l)
	}
	p := page.PcGts{
		Attributes: page.PcGtsXMLHeader,
		Metadata: page.Metadata{
			"Creator": "GOCRD",
			"Created": time.Now().Format(time.RFC3339),
		},
	}
	p.Page.TextRegion = append(p.Page.TextRegion, r)
	return xml.NewEncoder(os.Stdout).Encode(p)
}

func textRegionString(ocrlines []ocrline) string {
	var lines []string
	for _, ocrline := range ocrlines {
		lines = append(lines, ocrline.str)
	}
	return strings.Join(lines, "\n")
}

func imageRectangle(ocrlines []ocrline) []image.Point {
	var min, max image.Point
	for i, ocrline := range ocrlines {
		if i == 0 {
			min = ocrline.bb.Min
			max = ocrline.bb.Max
			continue
		}
		if ocrline.bb.Min.X < min.X {
			min.X = ocrline.bb.Min.X
		}
		if ocrline.bb.Min.Y < min.Y {
			min.Y = ocrline.bb.Min.Y
		}
		if ocrline.bb.Max.X > max.X {
			max.X = ocrline.bb.Max.X
		}
		if ocrline.bb.Max.Y > max.Y {
			max.Y = ocrline.bb.Max.Y
		}
	}
	return []image.Point{min, max}
}

var l lev.Lev

func bestFit(gtlines []string, ocrline string) (string, int) {
	min, argmin := 100000, -1
	for i, gtline := range gtlines {
		d := l.EditDistance(gtline, ocrline)
		if d < min {
			min = d
			argmin = i
		}
	}
	return gtlines[argmin], min
}

type ocrline struct {
	str string
	bb  image.Rectangle
}

func readOCRLines(path string) ([]ocrline, error) {
	is, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	var needText bool
	var lines []ocrline
	s := hocr.NewScanner(is)
	for s.Scan() {
		node := s.Node()
		if needText {
			if text, ok := node.(hocr.Text); ok {
				lines[len(lines)-1].str = string(text)
				needText = false
				continue
			}
		}
		element, ok := node.(hocr.Element)
		if !ok || element.Class != hocr.ClassLine {
			needText = false
			continue
		}
		lines = append(lines, ocrline{bb: element.BBox()})
		needText = true
	}
	return lines, s.Err()
}

func readGTLines(path string) ([]string, error) {
	is, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer is.Close()
	var lines []string
	s := bufio.NewScanner(is)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, s.Err()
}
