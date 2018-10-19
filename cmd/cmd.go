package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCMD = &cobra.Command{
	Use:   "gocrd",
	Short: "gocrd bundles various tools for ocrd",
}

// Execute is the main entry point for the gocrd commands.
func Execute() {
	if err := rootCMD.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v", err)
		os.Exit(1)
	}
}
