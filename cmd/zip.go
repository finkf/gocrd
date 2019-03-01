package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var zipCommand = &cobra.Command{
	Args:  cobra.ExactArgs(0),
	Use:   "zip",
	Short: "Zip a stream.",
	Long: `Zip a stream.

The first half of the input lines are zipped together with the second
half of the input lines.  Zip first reads all input lines into memory
before it can write them out.`,
	RunE: runZip}

var zipArgs struct {
	delim string
}

func init() {
	zipCommand.Flags().StringVarP(
		&zipArgs.delim, "delimiter", "d", "\n",
		"set the delimiter used to separate the zipped lines")
}

func runZip(cmd *cobra.Command, args []string) error {
	return zip(os.Stdin, os.Stdout)
}

func zip(in io.Reader, out io.Writer) error {
	var lines []string
	s := bufio.NewScanner(in)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if s.Err() != nil {
		return fmt.Errorf("read lines: %v", s.Err())
	}
	n := len(lines) / 2
	for i := 0; i < n; i++ {
		if err := printZip(out, lines[i], lines[i+n]); err != nil {
			return fmt.Errorf("write lines: %v", err)
		}
	}
	return nil
}

func printZip(out io.Writer, a, b string) error {
	_, err := fmt.Fprintf(out, "%s%s%s\n", a, zipArgs.delim, b)
	return err
}
