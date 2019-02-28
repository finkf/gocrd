package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/finkf/gocrd/xml/mets"
	"github.com/finkf/gocrd/xml/page"
	"github.com/spf13/cobra"
)

var catCommand = &cobra.Command{
	Use:   "cat",
	Short: "Print and concatenate page XML formatted files.",
	Long: `Print and concatenate page XML formatted files.

You can either specify the files as additional command line arguments
or as input file groups.  If input file groups are used, the according
mets file must be specified (thedefault can be used, though).  The
files of the input file groups are printed in sorted order of their
file names and in order of the specified input file groups given on
the command line.

If both input files and input file groups are given, first the input
files, then the input file groups are handled.`,
	RunE: runCat,
}

var catArgs struct {
	iffs, ifgs  []string // input files and input file groups
	level, mets string
	printHeader bool
}

func init() {
	catCommand.PersistentFlags().StringVarP(
		&catArgs.mets, "mets", "m", "mets.xml", "path to the METS file")
	catCommand.PersistentFlags().StringArrayVarP(
		&catArgs.ifgs, "input-file-grp", "I", nil, "input file groups")
	catCommand.Flags().StringVarP(
		&catArgs.level, "level", "l", "line", "set level of output regions [region|line|word]")
	catCommand.Flags().BoolVarP(
		&catArgs.printHeader, "header", "H", false, "ouput region id header")
}

func runCat(cmd *cobra.Command, args []string) error {
	if err := checkLevel(); err != nil {
		return err
	}
	catArgs.iffs = args
	return cat(os.Stdout)
}

func cat(out io.Writer) error {
	for _, iff := range catArgs.iffs {
		if err := catInputFile(out, iff); err != nil {
			return fmt.Errorf("input file %s: %v", iff, err)
		}
	}
	for _, ifg := range catArgs.ifgs {
		if err := catInputFileGroup(out, ifg); err != nil {
			return fmt.Errorf("input file group %s: %v", ifg, err)
		}
	}
	return nil
}

func catInputFile(out io.Writer, path string) error {
	iff, err := os.Open(path)
	if err != nil {
		fmt.Errorf("open %s: %v", path, err)
	}
	defer iff.Close()
	p, err := page.Read(iff)
	if err != nil {
		return fmt.Errorf("read %s: %v", path, err)
	}
	return catRegions(out, p)
}

func catInputFileGroup(out io.Writer, ifg string) error {
	m, err := mets.Open(catArgs.mets)
	if err != nil {
		return fmt.Errorf("open METS file %s: %v", catArgs.mets, err)
	}

	flocats := m.Find(mets.Match{Use: ifg}) // wuah-haha
	for _, file := range flocats {
		r, err := file.FLocat.Open()
		if err != nil {
			return fmt.Errorf("open FLocat %s: %v", file.FLocat.URL, err)
		}
		defer r.Close()
		p, err := page.Read(r)
		if err != nil {
			return fmt.Errorf("read %s: %v", file.FLocat.URL, err)
		}
		if err := catRegions(out, p); err != nil {
			return err
		}
	}
	return nil
}

func catRegions(out io.Writer, p *page.PcGts) error {
	level := strings.ToLower(catArgs.level)
	for _, region := range p.Page.TextRegion {
		if level == "region" {
			if err := catPrint(out, region.TextRegionBase); err != nil {
				return err
			}
			continue
		}
		for _, line := range region.TextLine {
			if level == "line" {
				if err := catPrint(out, line.TextRegionBase); err != nil {
					return err
				}
				continue
			}
			for _, word := range line.Word {
				if err := catPrint(out, word.TextRegionBase); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func catPrint(out io.Writer, tr page.TextRegionBase) error {
	if catArgs.printHeader {
		if _, err := fmt.Fprintln(out, tr.ID); err != nil {
			return err
		}
	}
	var str string
	if len(tr.TextEquiv.Unicode) > 0 {
		str = tr.TextEquiv.Unicode[0]
	}
	_, err := fmt.Fprintln(out, str)
	return err
}

func checkLevel() error {
	switch strings.ToLower(catArgs.level) {
	case "line":
		return nil
	case "region":
		return nil
	case "word":
		return nil
	}
	return fmt.Errorf("invalid level: %s", catArgs.level)
}
