package hocr

import (
	"fmt"
	"testing"
)

func TestStackPushPop(t *testing.T) {
	tests := []struct {
		test []string
		want string
	}{
		{[]string{"push", "push"}, "[push push]"},
		{[]string{"push", "push", "pop"}, "[push]"},
		{[]string{"push", "push", "pop", "pop"}, "[]"},
		{[]string{"push", "pop", "push", "pop"}, "[]"},
		{[]string{"push", "pop", "push"}, "[push]"},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.test), func(t *testing.T) {
			var s stack
			for _, op := range tc.test {
				if op == "push" {
					s = s.push(op)
				} else {
					s = s.pop()
				}
			}
			if got := fmt.Sprintf("%v", s); got != tc.want {
				t.Fatalf("expected %s; got %s", tc.want, got)
			}
		})
	}
}

func TestStackMatch(t *testing.T) {
	tests := []struct {
		stack stack
		test  []string
		want  bool
	}{
		{stack{"a", "b", "c"}, []string{"c"}, true},
		{stack{"a", "b", "c"}, []string{"b", "c"}, true},
		{stack{"a", "b", "c"}, []string{"a", "b", "c"}, true},
		{stack{"a", "b", "c"}, []string{"a", "a", "b", "c"}, false},
		{stack{"a", "b", "c"}, []string{"a", "a", "b"}, false},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.test), func(t *testing.T) {
			if got := tc.stack.match(tc.test...); got != tc.want {
				t.Fatalf("expected %t; got %t", tc.want, got)
			}
		})
	}
}
