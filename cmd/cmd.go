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
	metsFile        string
	inputFileGroups []string
)

func init() {
	rootCMD.PersistentFlags().StringVarP(
		&metsFile, "mets", "m", "mets.xml", "path to the workspace's mets file")
	rootCMD.PersistentFlags().StringArrayVarP(
		&inputFileGroups, "input-file-grp", "I", nil, "input file groups")

	rootCMD.AddCommand(catCMD)
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
