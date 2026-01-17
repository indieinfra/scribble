package util

import (
	"strings"
	"testing"
)

func TestHtmlToText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		words    int
		expected string
	}{
		{
			name:     "simple html",
			input:    "<p>Hello world</p>",
			words:    10,
			expected: "Hello world",
		},
		{
			name:     "html with multiple tags",
			input:    "<div><p>This is</p><p>a test</p></div>",
			words:    10,
			expected: "This is a test",
		},
		{
			name:     "truncate to first word",
			input:    "<p>One two three four five</p>",
			words:    1,
			expected: "One",
		},
		{
			name:     "truncate to first three words",
			input:    "<p>The quick brown fox jumps</p>",
			words:    3,
			expected: "The quick brown",
		},
		{
			name:     "fewer words than requested",
			input:    "<p>Just two</p>",
			words:    10,
			expected: "Just two",
		},
		{
			name:     "plain text input",
			input:    "Hello world from plain text",
			words:    3,
			expected: "Hello world from",
		},
		{
			name:     "empty input",
			input:    "",
			words:    5,
			expected: "",
		},
		{
			name:     "zero words requested",
			input:    "<p>Hello world</p>",
			words:    0,
			expected: "",
		},
		{
			name:     "negative words requested",
			input:    "<p>Hello world</p>",
			words:    -1,
			expected: "",
		},
		{
			name:     "html with extra whitespace",
			input:    "<p>Hello   world   test</p>",
			words:    3,
			expected: "Hello world test",
		},
		{
			name:     "nested html with text",
			input:    "<div><span>Nested <strong>text</strong> here</span></div>",
			words:    10,
			expected: "Nested text here",
		},
		{
			name:     "html with script tag (should be ignored)",
			input:    "<p>Hello</p><script>alert('ignored')</script><p>world</p>",
			words:    10,
			expected: "Hello world",
		},
		{
			name:     "html with style tag (should be ignored)",
			input:    "<p>Hello</p><style>.ignored { color: red; }</style><p>world</p>",
			words:    10,
			expected: "Hello world",
		},
		{
			name:     "malformed html treated as plain text",
			input:    "This < is not > really html",
			words:    4,
			expected: "This < is not",
		},
		{
			name:     "only whitespace",
			input:    "   \n\t   ",
			words:    5,
			expected: "",
		},
		{
			name:     "html with newlines and tabs",
			input:    "<p>Hello\n\tworld\n\ttest</p>",
			words:    10,
			expected: "Hello world test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HtmlToText(tt.input, tt.words)
			if result != tt.expected {
				t.Errorf("HtmlToText(%q, %d) = %q, want %q", tt.input, tt.words, result, tt.expected)
			}
		})
	}
}

// FuzzHtmlToText tests HtmlToText with random inputs to catch panics and unexpected behavior
func FuzzHtmlToText(f *testing.F) {
	// Add seed corpus for better coverage
	seedInputs := []struct {
		input string
		words int
	}{
		{"<p>Hello world</p>", 2},
		{"Hello world", 1},
		{"", 0},
		{"<>", 1},
		{"<p>", 1},
		{"<script>alert('test')</script>", 5},
		{"   \n\t   ", 3},
		{"a b c d e f g h i j", 5},
	}

	for _, seed := range seedInputs {
		f.Add(seed.input, seed.words)
	}

	f.Fuzz(func(t *testing.T, input string, words int) {
		// The function should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("HtmlToText panicked with input %q, words %d: %v", input, words, r)
			}
		}()

		result := HtmlToText(input, words)

		// Count words in result to ensure it doesn't exceed requested count
		resultWords := len(strings.Fields(result))
		maxWords := words
		if maxWords < 0 {
			maxWords = 0
		}
		if resultWords > maxWords {
			t.Errorf("HtmlToText(%q, %d) returned %d words, expected at most %d: %q", input, words, resultWords, maxWords, result)
		}

		// Result should not contain control characters (they should be stripped)
		for _, c := range result {
			if c < 32 && c != ' ' && c != '\t' && c != '\n' && c != '\r' {
				t.Errorf("HtmlToText(%q, %d) returned unexpected control character: %d", input, words, c)
			}
			if c == 127 { // DEL character
				t.Errorf("HtmlToText(%q, %d) returned unexpected DEL character", input, words)
			}
		}
	})
}
