package cmd

import (
	"bufio"
	"encoding/xml"
	"os"
	"strings"

	"github.com/finkf/gocrd/xml/page"
	"github.com/finkf/lev"
	"github.com/spf13/cobra"
)

var replaceCommand = &cobra.Command{
	Use:   "replace txt page-xml output",
	Short: "replace text in a PageXML file",
	Long: `Given a text file and a PageXML file replace
all non empty lines in the text file with the
non empty lines in the PageXML file.  Lines are
aligned and unalignable lines in the PageXML
file are discarded.  No new lines are introduced
in the PageXML file.  The resulting file is written
to the output file.`,
	Args: cobra.ExactArgs(3),
	RunE: runReplace,
}

func runReplace(cmd *cobra.Command, args []string) error {
	return replace(args[0], args[1], args[2])
}

func replace(txt, pagexml, output string) error {
	replaceLines, err := readReplacementLines(txt)
	if err != nil {
		return err
	}
	p, err := page.Open(pagexml)
	if err != nil {
		return err
	}
	if err := replacePage(p, replaceLines); err != nil {
		return err
	}
	return writePage(p, output)
}

func replacePage(p *page.PcGts, replaceLines []string) error {
	var pageLines []string
	var pageLineIndices []struct{ r, l int }
	for r, region := range p.Page.TextRegion {
		for l, line := range region.TextLine {
			if len(line.TextEquiv.Unicode) == 0 {
				continue
			}
			line := strings.Trim(line.TextEquiv.Unicode[0], "\t\n\r\v ")
			if line == "" {
				continue
			}
			pageLines = append(pageLines, line)
			pageLineIndices = append(pageLineIndices, struct{ r, l int }{r, l})
		}
	}
	var l lev.Lev
	var w lev.WLev
	a := lev.NewStringArray(&l, replaceLines...)
	b := lev.NewStringArray(&l, pageLines...)
	_, t := w.Trace(a, b)

	lev.Align(t, func(op byte, i, j int) {
		switch op {
		case lev.Nop:
			// do nothing
		case lev.Sub:
			substituteLine(p, replaceLines[i], pageLineIndices[j])
		case lev.Del:
			deleteLine(p, pageLineIndices[i])
		case lev.Ins:
			deleteLine(p, pageLineIndices[j])
		}
	})
	return nil
}

func substituteLine(p *page.PcGts, new string, i struct{ r, l int }) {
	line := &p.Page.TextRegion[i.r].TextLine[i.l]
	line.TextEquiv.Unicode[0] = new
	line.UpdateWords(line.TextEquiv.Unicode[0])
}

func deleteLine(p *page.PcGts, i struct{ r, l int }) {
	// append(slice[:s], slice[s+1:]...)
	p.Page.TextRegion[i.r].TextLine = append(
		p.Page.TextRegion[i.r].TextLine[:i.l],
		p.Page.TextRegion[i.r].TextLine[i.l+1:]...,
	)
}

func writePage(p *page.PcGts, output string) (oerr error) {
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() { oerr = checkClose(oerr, out) }()
	e := xml.NewEncoder(out)
	e.Indent("\t", "\t")
	return e.Encode(p)
}

func readReplacementLines(input string) ([]string, error) {
	in, err := os.Open(input)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	var strs []string
	s := bufio.NewScanner(in)
	for s.Scan() {
		text := strings.Trim(s.Text(), "\t\n\r\v ")
		if text == "" {
			continue
		}
		strs = append(strs, text)
	}
	return strs, s.Err()
}
