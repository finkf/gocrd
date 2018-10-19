package cmd

import (
	"fmt"
	"strings"

	"github.com/finkf/gocrd/mets"
	"github.com/finkf/gocrd/page"
	"github.com/spf13/cobra"
)

var catCMD = &cobra.Command{
	Use:   "cat",
	Short: "concatenates files and prints them to stdout",
	RunE:  cat,
}
var catLevel string

func init() {
	catCMD.Flags().StringVarP(
		&catLevel, "level", "l", "line", "set level of output regions [region|line|word]")
}

func cat(cmd *cobra.Command, args []string) error {
	if !levelOK() {
		return fmt.Errorf("invalid level: %s", catLevel)
	}
	m, err := mets.Open(metsFile)
	if err != nil {
		return fmt.Errorf("cannot open mets file: %v", err)
	}
	for _, ifg := range inputFileGroups {
		for _, f := range m.Find(mets.Match{Use: ifg}) {
			if err := catFLocat(f.FLocat); err != nil {
				return fmt.Errorf("cannot open %s: %v", f.FLocat.URL, err)
			}
		}
	}
	return nil
}

func catFLocat(f mets.FLocat) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	p, err := page.Parse(r, f.URL)
	if err != nil {
		return err
	}
	p.EachSubRegion(func(r page.TextRegion) {
		if strings.ToLower(catLevel) == "region" {
			printRegion(r)
			return
		}
		r.EachSubRegion(func(r page.TextRegion) {
			if strings.ToLower(catLevel) == "line" {
				printRegion(r)
				return
			}
			r.EachSubRegion(func(r page.TextRegion) {
				printRegion(r)
			})
		})
	})
	return nil
}

func printRegion(r page.TextRegion) {
	line, ok := r.TextEquivUnicodeAt(0)
	if !ok {
		return
	}
	fmt.Printf("%s\n", strings.Replace(line, "\n", " ", -1))
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
