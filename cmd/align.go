package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/finkf/lev"
	"github.com/spf13/cobra"
)

var alignCommand = cobra.Command{
	Use:   "align",
	Short: `Align pairs of input lines.`,
	Long: `Align pairs of input lines.

Read consecutive pairs of lines, align them and print out the
alignment patterns.`,
	RunE: runAlign,
	Args: cobra.ExactArgs(0),
}

var alignArgs struct {
	header bool
}

func init() {
	alignCommand.Flags().BoolVarP(&alignArgs.header, "header", "H",
		false, "read an additional filename as first line from the input")
}

func runAlign(cmd *cobra.Command, args []string) error {
	return align(os.Stdin, os.Stdout)
}

func align(in io.Reader, out io.Writer) error {
	var l lev.Lev
	s := newTripleScanner(in, alignArgs.header)
	for s.scan() {
		h, a, b := s.triple()
		al, err := l.Alignment(l.Trace(a, b))
		if err != nil {
			fmt.Errorf("alignment: %v", err)
		}
		if err := alignPrint(out, h, al); err != nil {
			return fmt.Errorf("write alignment: %v", err)
		}
	}
	return s.err()
}

func alignPrint(out io.Writer, header string, a lev.Alignment) error {
	if header != "" {
		if _, err := fmt.Fprintln(out, header); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(out, "%s\n%s\n%s\n",
		string(a.S1), a.Trace, string(a.S2))
	return err
}

type tripleScanner struct {
	e   error
	s   *bufio.Scanner
	buf [3]string // buffer for header, first and second string (in this order)
	o   int       // offset into buf
}

func newTripleScanner(r io.Reader, header bool) *tripleScanner {
	s := &tripleScanner{s: bufio.NewScanner(r)}
	if !header {
		s.o = 1 // skip first entry if no header needs to be parsed
	}
	return s
}

func (s *tripleScanner) scan() bool {
	if s.e != nil || !s.s.Scan() { // EOF or error; we are done
		return false
	}
	s.buf[s.o] = s.s.Text() // header or first string depending on offset
	for i := s.o + 1; i < len(s.buf); i++ {
		if !s.s.Scan() {
			if s.s.Err() != nil {
				return false
			}
			s.e = fmt.Errorf("premature EOF")
			return false
		}
		s.buf[i] = s.s.Text()
	}
	return true
}

func (s *tripleScanner) triple() (h, a, b string) {
	return s.buf[0], s.buf[1], s.buf[2]
}

func (s *tripleScanner) err() error {
	if s.e != nil {
		return s.e
	}
	return s.s.Err()
}
