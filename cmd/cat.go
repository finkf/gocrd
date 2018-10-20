package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/finkf/gocrd/mets"
	"github.com/finkf/gocrd/page"
	"github.com/spf13/cobra"
)

var catCMD = &cobra.Command{
	Use:   "cat",
	Short: "concatenates two input file groups and prints them to stdout",
	RunE:  cat,
}
var (
	catLevel    string
	printHeader bool
	catOut      io.Writer = os.Stdout
)

func init() {
	catCMD.Flags().StringVarP(
		&catLevel, "level", "l", "line", "set level of output regions [region|line|word]")
	catCMD.Flags().BoolVarP(
		&printHeader, "header", "H", false, "ouput region id header")
}

func cat(cmd *cobra.Command, args []string) error {
	if !levelOK() {
		return fmt.Errorf("invalid level: %s", catLevel)
	}
	if !inputGroupsHaveLen2() {
		return fmt.Errorf("expected two input file groups (%d given)", len(inputFileGroups))
	}
	m, err := mets.Open(metsFile)
	if err != nil {
		return fmt.Errorf("cannot open mets file: %v", err)
	}
	ifg1 := m.Find(mets.Match{Use: inputFileGroups[0]})
	ifg2 := m.Find(mets.Match{Use: inputFileGroups[1]})
	err = zip(ifg1, ifg2, func(a, b mets.File) error {
		ls1, e2 := readRegions(a.FLocat)
		if e2 != nil {
			return fmt.Errorf("cannot read %s: %v", a.FLocat.URL, err)
		}
		ls2, e2 := readRegions(b.FLocat)
		if e2 != nil {
			return fmt.Errorf("cannot read %s: %v", b.FLocat.URL, err)
		}
		for i := 0; i < len(ls1); i += 2 {
			if printHeader {
				fmt.Fprintf(catOut, "%s\n", ls1[i])
			}
			fmt.Fprintf(catOut, "%s\n%s", ls1[i+1], ls2[i+1])
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
	p, err := page.Parse(r, f.URL)
	if err != nil {
		return nil, err
	}
	var res []string
	p.EachSubRegion(func(r page.TextRegion) {
		if strings.ToLower(catLevel) == "region" {
			res = append(res, r.ID(), regionString(r))
			return
		}
		r.EachSubRegion(func(r page.TextRegion) {
			if strings.ToLower(catLevel) == "line" {
				res = append(res, r.ID(), regionString(r))
				return
			}
			r.EachSubRegion(func(r page.TextRegion) {
				res = append(res, r.ID(), regionString(r))
			})
		})
	})
	return res, nil
}

func regionString(r page.TextRegion) string {
	line, ok := r.TextEquivUnicodeAt(0)
	if !ok {
		return ""
	}
	return strings.Replace(line, "\n", " ", -1)
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
