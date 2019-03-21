package boundingbox

import (
	"image"
	"strings"
	"unicode"
)

// Split represents (trimmed) words with their accoring right
// position in a rectangle.
type Split struct {
	Str string
	Cut int
}

// ToPoints converts a rectangle to an array of two points, consisting
// of the top left point (rectangle.Min) and the bottom right point
// (rectangle.Max).
func ToPoints(rect image.Rectangle) []image.Point {
	return []image.Point{rect.Min, rect.Max}
}

// SplitTokens splits a given rectangle into an list of tokens and
// their approixmate right cuts.  Whitespace between tokens is
// distributed between the tokens.
func SplitTokens(rect image.Rectangle, str string) []Split {
	if str == "" {
		return nil
	}
	var splits []Split
	wstr := []rune(str)
	cuts := Cuts(rect, len(wstr))
	for b, i := 0, 0; i < len(wstr); i++ {
		// skip leading whitespace
		for i < len(wstr) && unicode.IsSpace(wstr[i]) {
			i++
		}
		if i >= len(wstr) {
			break
		}
		// find end of token
		for i < len(wstr) && !unicode.IsSpace(wstr[i]) {
			i++
		}
		var cut int
		if i < len(cuts) {
			cut = cuts[i]
		} else {
			cut = cuts[len(cuts)-1]
		}
		splits = append(splits, Split{
			Str: strings.Trim(string(wstr[b:i]), "\t\n\r\v "),
			Cut: cut,
		})
		b = i + 1
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
