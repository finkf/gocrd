package boundingbox

import (
	"fmt"
	"image"
	"strings"
)

type Split struct {
	Str string
	Cut int
}

// SplitTokens splits a given rectangle into an list of
// tokens and their approixmate right cuts.
func SplitTokens(rect image.Rectangle, str string) []Split {
	fields := strings.Fields(str)
	splits := make([]Split, len(fields))
	for _, _ = range strings.Fields(str) {
	}
	return splits
}

// Cuts returns an array of evenly spaced right positions for n
// characters within the given rectangle.  If n is larger then the
// rectangle's width, this function panics.
func Cuts(rect image.Rectangle, n int) []int {
	if n > rect.Dx() {
		panic(fmt.Sprintf("cannot calculate %d cuts for rectangle of width %d", n, rect.Dx()))
	}
	w := rect.Dx() / n // width for each cut
	r := rect.Dx() % n // rest to distribute
	b := rect.Min.X    // left position
	cuts := make([]int, n)
	for i := 0; i < n; i++ {
		cuts[i] = b + w
		if r > 0 {
			cuts[i] += 1
			r--
		}
		b = cuts[i]
	}
	return cuts
}
