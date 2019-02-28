package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
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
	rootCommand.AddCommand(catCommand)
	rootCommand.AddCommand(convertCommand)
	rootCommand.AddCommand(zipCommand)
}

// Execute is the main entry point for the gocrd commands.
func Execute() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err)
		os.Exit(1)
	}
}
