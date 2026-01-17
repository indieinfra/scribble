package util

import "testing"

func TestFirstN(t *testing.T) {
	cases := []struct {
		name string
		s    []int
		n    int
		exp  []int
	}{
		{"exact length", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"shorter", []int{1, 2, 3}, 2, []int{1, 2}},
		{"longer than slice", []int{1, 2}, 5, []int{1, 2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FirstN(tc.s, tc.n)
			if len(got) != len(tc.exp) {
				t.Fatalf("len mismatch: got %d want %d", len(got), len(tc.exp))
			}
			for i := range got {
				if got[i] != tc.exp[i] {
					t.Fatalf("at %d got %d want %d", i, got[i], tc.exp[i])
				}
			}
		})
	}
}

func TestFirstNWords(t *testing.T) {
	cases := []struct {
		name  string
		input string
		n     int
		exp   []string
	}{
		{"basic", "one two three four", 2, []string{"one", "two"}},
		{"longer n", "one two", 5, []string{"one", "two"}},
		{"empty", "", 3, []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FirstNWords(tc.input, tc.n)
			if len(got) != len(tc.exp) {
				t.Fatalf("len mismatch: got %d want %d", len(got), len(tc.exp))
			}
			for i := range got {
				if got[i] != tc.exp[i] {
					t.Fatalf("at %d got %q want %q", i, got[i], tc.exp[i])
				}
			}
		})
	}
}
