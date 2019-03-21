package boundingbox

import (
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"
)

func TestToPoints(t *testing.T) {
	tests := []struct {
		test image.Rectangle
		want []image.Point
	}{
		{image.Rect(1, 1, 2, 2), []image.Point{{1, 1}, {2, 2}}},
		{image.Rect(0, 0, 0, 0), []image.Point{{0, 0}, {0, 0}}},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s", tc.test), func(t *testing.T) {
			if got := ToPoints(tc.test); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %v; got %v", tc.want, got)
			}
		})
	}
}

func TestFromPoints(t *testing.T) {
	tests := []struct {
		test []image.Point
		want image.Rectangle
	}{
		{nil, image.Rect(0, 0, math.MaxInt64, math.MaxInt64)},
		{[]image.Point{{1, 1}}, image.Rect(1, 1, 1, 1)},
		{[]image.Point{{1, 1}, {2, 2}}, image.Rect(1, 1, 2, 2)},
		{[]image.Point{{1, 1}, {2, 2}, {3, 3}}, image.Rect(1, 1, 3, 3)},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.test), func(t *testing.T) {
			if got := FromPoints(tc.test); got != tc.want {
				t.Fatalf("expected %s; got %s", tc.want, got)
			}
		})
	}
}

func TestCuts(t *testing.T) {
	tests := []struct {
		want []int
		test image.Rectangle
		n    int
	}{
		{[]int{1, 1, 1}, image.Rect(0, 0, 1, 0), 3},
		{[]int{1, 2, 2}, image.Rect(0, 0, 2, 0), 3},
		{[]int{1, 2, 3}, image.Rect(0, 0, 3, 0), 3},
		{[]int{2, 3, 4}, image.Rect(0, 0, 4, 0), 3},
		{[]int{2, 4, 5}, image.Rect(0, 0, 5, 0), 3},
		{[]int{2, 4, 6}, image.Rect(0, 0, 6, 0), 3},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s-%d", tc.test, tc.n), func(t *testing.T) {
			if got := Cuts(tc.test, tc.n); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %v; got %v", tc.want, got)
			}
		})
	}
}

func TestSplitTokens(t *testing.T) {
	tests := []struct {
		want []Split
		test string
		rect image.Rectangle
	}{
		{nil, "", image.Rect(0, 0, 3, 0)},
		{nil, "  ", image.Rect(0, 0, 3, 0)},
		{[]Split{{"a", 2}, {"b", 3}}, "a b", image.Rect(0, 0, 3, 0)},
		{[]Split{{"a", 3}, {"b", 4}}, " a b ", image.Rect(0, 0, 4, 0)},
		{[]Split{{"a", 3}, {"b", 5}}, " a  b ", image.Rect(0, 0, 5, 0)},
		{[]Split{{"a", 3}, {"b", 5}}, " a  b  ", image.Rect(0, 0, 5, 0)},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			if got := SplitTokens(tc.rect, tc.test); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %v; got %v", tc.want, got)
			}
		})
	}
}
