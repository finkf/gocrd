package boundingbox

import (
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
// positions within the given rectangle.  N must be larger than 0.
func Cuts(rect image.Rectangle, n int) []int {
	w := rect.Dx() / n // width for each cut
	r := rect.Dx() % n // rest to distribute
	b := rect.Min.X    // left position
	cuts := make([]int, n)
	for i := 0; i < n; i++ {
		cuts[i] = b + w
		if r > 0 {
			cuts[i]++
			r--
		}
		b = cuts[i]
	}
	return cuts
}
