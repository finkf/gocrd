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

var catCMD = &cobra.Command{
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
var (
	catLevel    string
	printHeader bool
)

func init() {
	catCMD.Flags().StringVarP(
		&catLevel, "level", "l", "line", "set level of output regions [region|line|word]")
	catCMD.Flags().BoolVarP(
		&printHeader, "header", "H", false, "ouput region id header")
}

func runCat(cmd *cobra.Command, args []string) error {
	if !levelOK() {
		return fmt.Errorf("invalid level: %s", catLevel)
	}
	if !inputGroupsHaveLen2() {
		return fmt.Errorf("expected two input file groups (%d given)", len(inputFileGroups))
	}
	return cat(os.Stdout, catArgs{
		mets:        metsFile,
		level:       catLevel,
		ifg1:        inputFileGroups[0],
		ifg2:        inputFileGroups[1],
		printHeader: printHeader,
	})
}

type catArgs struct {
	mets, level, ifg1, ifg2 string
	printHeader             bool
}

func cat(out io.Writer, args catArgs) error {
	m, err := mets.Open(args.mets)
	if err != nil {
		return fmt.Errorf("cannot open mets file: %v", err)
	}
	ifg1 := m.Find(mets.Match{Use: args.ifg1})
	ifg2 := m.Find(mets.Match{Use: args.ifg2})
	err = zip(ifg1, ifg2, func(a, b mets.File) error {
		ls1, e2 := readRegions(a.FLocat)
		if e2 != nil {
			return fmt.Errorf("cannot read %s: %v", a.FLocat.URL, e2)
		}
		ls2, e2 := readRegions(b.FLocat)
		if e2 != nil {
			return fmt.Errorf("cannot read %s: %v", b.FLocat.URL, e2)
		}
		for i := 0; i < len(ls1); i += 2 {
			if args.printHeader {
				fmt.Fprintf(out, "%s\n", ls1[i])
			}
			fmt.Fprintf(out, "%s\n%s", ls1[i+1], ls2[i+1])
		}
		return nil
	})
	return err
}

func readRegions(f mets.FLocat) ([]string, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	p, err := page.Read(r)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, region := range p.Page.TextRegion {
		if strings.ToLower(catLevel) == "region" {
			res = append(res, region.ID, regionString(region.TextEquiv))
			continue
		}
		for _, line := range region.TextLine {
			if strings.ToLower(catLevel) == "line" {
				res = append(res, line.ID, regionString(line.TextEquiv))
				continue
			}
			for _, word := range line.Word {
				res = append(res, word.ID, regionString(word.TextEquiv))
			}
		}
	}
	return res, nil
}

func regionString(t page.TextEquiv) string {
	if len(t.Unicode) == 0 {
		return ""
	}
	return strings.Replace(t.Unicode[0], "\n", " ", -1)
}

func levelOK() bool {
	switch strings.ToLower(catLevel) {
	case "line":
		return true
	case "region":
		return true
	case "word":
		return true
	default:
		return false
	}
}

func inputGroupsHaveLen2() bool {
	return len(inputFileGroups) == 2
}

func zip(a, b []mets.File, f func(mets.File, mets.File) error) error {
	for i := 0; i < min(len(a), len(b)); i++ {
		if err := f(a[i], b[i]); err != nil {
			return err
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
