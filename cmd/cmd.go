package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var rootCMD = &cobra.Command{
	Use:           "gocrd",
	Short:         "gocrd bundles various tools for ocrd",
	SilenceUsage:  true,
	SilenceErrors: true,
}
var (
	metsFile      string
	inputFileGrps []string
)

const (
	levelRegion = "region"
	levelLine   = "line"
	levelWord   = "word"
	levelGlyph  = "glyph"
)

func init() {
	rootCMD.AddCommand(catCommand)
	rootCMD.AddCommand(convertCommand)
	rootCMD.AddCommand(replaceCommand)
	rootCMD.AddCommand(synpageCommand)
	rootCMD.AddCommand(markCommand)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// checkClose helps to defer close calls with error checking.
func checkClose(err error, c io.Closer) error {
	errClose := c.Close()
	if err != nil {
		return err
	}
	return errClose
}

// Execute is the main entry point for the gocrd commands.
func Execute() {
	if err := rootCMD.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}
}
