package util

import "testing"

func TestPopAccessToken(t *testing.T) {
	cases := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: "",
		},
		{
			name:     "no token",
			input:    map[string]any{"foo": "bar"},
			expected: "",
		},
		{
			name:     "string token",
			input:    map[string]any{"access_token": "abc"},
			expected: "abc",
		},
		{
			name:     "slice token picks first string",
			input:    map[string]any{"access_token": []any{"", "token2", "ignored"}},
			expected: "token2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := PopAccessToken(tc.input)
			if got != tc.expected {
				t.Fatalf("PopAccessToken() = %q, want %q", got, tc.expected)
			}

			if tc.input != nil {
				if _, exists := tc.input["access_token"]; exists {
					t.Fatalf("access_token was not removed from map")
				}
			}
		})
	}
}
