package cmd

import (
	"fmt"
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

func init() {
	rootCMD.AddCommand(catCommand)
	rootCMD.AddCommand(convertCommand)
	rootCMD.AddCommand(replaceCommand)
	rootCMD.AddCommand(synpageCommand)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// Execute is the main entry point for the gocrd commands.
func Execute() {
	if err := rootCMD.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}
}
