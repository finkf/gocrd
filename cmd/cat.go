package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/finkf/gocrd/xml/mets"
	"github.com/finkf/gocrd/xml/page"
	"github.com/spf13/cobra"
)

var catCommand = &cobra.Command{
	Use:   "cat [flags] [files...]",
	Short: "concatenates the contents of PageXML files to stdout",
	RunE:  runCatFunc,
}

const catDefaultFormat = "%R:"

var (
	catFormat             = catDefaultFormat
	catLevel              string
	catPrefix             bool
	catFormatFileID       string
	catFormatInputFileGrp string
	catFormatRegionID     string
	catFormatN            int
)

func init() {
	catCommand.Flags().StringVarP(
		&metsFile, "mets", "m", "mets.xml", "path to the workspace's mets file")
	catCommand.Flags().StringArrayVarP(
		&inputFileGrps, "input-file-grp", "I", nil, "input file groups")
	catCommand.Flags().StringVarP(
		&catLevel, "level", "l", "line", "set level of output regions [region|line|word]")
	catCommand.Flags().BoolVarP(
		&catPrefix, "prefix", "p", false, "ouput default prefix")
	catCommand.Flags().StringVarP(
		&catFormat, "format", "f", catFormat, "set formatting for prefix")
}

func runCatFunc(cmd *cobra.Command, args []string) error {
	return runCat(args, os.Stdout)
}

func runCat(args []string, stdout io.Writer) error {
	if !levelOK() {
		return fmt.Errorf("invalid level: %s", catLevel)
	}
	for _, ifg := range inputFileGrps {
		if err := catInputFileGrp(ifg, stdout); err != nil {
			return err
		}
	}
	for _, arg := range args {
		if err := catFile(arg, stdout); err != nil {
			return err
		}
	}
	return nil
}

func catFile(p string, stdout io.Writer) error {
	is, err := os.Open(p)
	if err != nil {
		return err
	}
	defer is.Close()
	catFormatFileID = p
	catFormatInputFileGrp = path.Dir(p)
	return cat(is, stdout)
}

func catInputFileGrp(ifg string, stdout io.Writer) error {
	m, err := mets.Open(metsFile)
	if err != nil {
		return err
	}
	catFormatInputFileGrp = ifg
	for _, file := range m.FindFileGrp(ifg) {
		is, err := m.Open(file)
		if err != nil {
			return err
		}
		defer is.Close()
		catFormatFileID = file.ID
		if err = cat(is, stdout); err != nil {
			return err
		}
	}
	return nil
}

func cat(is io.Reader, stdout io.Writer) error {
	p, err := page.Read(is)
	if err != nil {
		return err
	}
	return catPage(p.Page, stdout)
}

func catPage(p page.Page, stdout io.Writer) error {
	// cat level is either region, word or glyph
	pregion := strings.ToLower(catLevel) == "region"
	pline := strings.ToLower(catLevel) == "line"

	for _, region := range p.TextRegion {
		if pregion {
			catFormatRegionID = region.ID
			if err := catprint(stdout, regionString(region.TextEquiv)); err != nil {
				return err
			}
			continue
		}
		for _, line := range region.TextLine {
			if pline {
				catFormatRegionID = line.ID
				if err := catprint(stdout, regionString(line.TextEquiv)); err != nil {
					return err
				}
				continue
			}
			for _, word := range line.Word {
				catFormatRegionID = word.ID
				if err := catprint(stdout, regionString(word.TextEquiv)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func catprint(stdout io.Writer, line string) error {
	if catPrefix || catFormat != catDefaultFormat {
		return catprintf(stdout, line)
	}
	_, err := fmt.Fprintln(stdout, line)
	return err
}

var catPrefixBuilder strings.Builder

func catprintf(stdout io.Writer, line string) error {
	catFormatN++
	catPrefixBuilder.Reset()
	var haveFormat bool
	for _, r := range catFormat {
		if haveFormat {
			switch r {
			case 'N':
				catPrefixBuilder.WriteString(fmt.Sprintf("%d", catFormatN))
			case 'G':
				catPrefixBuilder.WriteString(catFormatInputFileGrp)
			case 'F':
				catPrefixBuilder.WriteString(catFormatFileID)
			case 'R':
				catPrefixBuilder.WriteString(catFormatRegionID)
			case '%':
				catPrefixBuilder.WriteRune('%')
			default:
				return fmt.Errorf("invalid format: %%%c", r)
			}
			haveFormat = false
			continue
		}
		if r == '%' {
			haveFormat = true
			continue
		}
		catPrefixBuilder.WriteRune(r)
	}
	_, err := fmt.Fprintf(stdout, "%s%s\n", catPrefixBuilder.String(), line)
	return err
}

func regionString(t page.TextEquiv) string {
	if len(t.Unicode) == 0 {
		return ""
	}
	return strings.Replace(t.Unicode[0], "\n", " ", -1)
}

func levelOK() bool {
	switch strings.ToLower(catLevel) {
	case "line", "region", "word":
		return true
	default:
		return false
	}
}
