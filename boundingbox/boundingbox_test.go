package boundingbox

import (
	"fmt"
	"image"
	"reflect"
	"testing"
)

func TestCuts(t *testing.T) {
	tests := []struct {
		want []int
		test image.Rectangle
		n    int
	}{
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

// func TestCuts(t *testing.T) {
// 	tests := []struct {
// 		want []int
// 		test image.Rectangle
// 		n    int
// 	}{
// 		{[]int{1, 2, 3}, image.Rect(0, 0, 3, 0), 3},
// 		{[]int{2, 3, 4}, image.Rect(0, 0, 4, 0), 3},
// 		{[]int{2, 4, 5}, image.Rect(0, 0, 5, 0), 3},
// 		{[]int{2, 4, 6}, image.Rect(0, 0, 6, 0), 3},
// 	}
// 	for _, tc := range tests {
// 		t.Run(fmt.Sprintf("%s-%d", tc.test, tc.n), func(t *testing.T) {
// 			if got := Cuts(tc.test, tc.n); !reflect.DeepEqual(got, tc.want) {
// 				t.Fatalf("expected %v; got %v", tc.want, got)
// 			}
// 		})
// 	}
// }
