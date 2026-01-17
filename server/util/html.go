package util

import (
	"strings"

	"golang.org/x/net/html"
)

func HtmlToText(input string, words int) string {
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		// If parsing fails, treat input as plain text
		return truncateWords(input, words)
	}

	// Extract text from parsed HTML, stopping once we have enough words
	text := extractText(doc, words)

	// Return first N words
	return truncateWords(text, words)
}

// extractText recursively extracts text content from HTML nodes, stopping once maxWords is reached
func extractText(n *html.Node, maxWords int) string {
	var text strings.Builder
	wordCount := 0

	var traverse func(*html.Node) bool
	traverse = func(n *html.Node) bool {
		// Skip script and style tags
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return false
		}

		if n.Type == html.TextNode {
			// Add text content with spaces
			content := strings.TrimSpace(n.Data)
			if content != "" {
				if text.Len() > 0 {
					text.WriteByte(' ')
				}
				text.WriteString(content)

				// Count words in the content we just added
				wordCount += len(strings.Fields(content))

				// Stop if we've reached the word limit
				if maxWords > 0 && wordCount >= maxWords {
					return true
				}
			}
		}

		// Traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if traverse(c) {
				return true // Stop traversal
			}
		}
		return false
	}

	traverse(n)
	return text.String()
}

// truncateWords returns the first N words from the input string
func truncateWords(text string, maxWords int) string {
	if maxWords <= 0 {
		return ""
	}

	// Remove control characters from text
	text = removeControlCharacters(text)

	// Split by whitespace, removing empty strings
	wordList := strings.Fields(text)

	if len(wordList) <= maxWords {
		return strings.Join(wordList, " ")
	}

	return strings.Join(wordList[:maxWords], " ")
}

// removeControlCharacters removes ASCII control characters from the input string
func removeControlCharacters(text string) string {
	return strings.Map(func(r rune) rune {
		// Keep printable characters, spaces, and common whitespace
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1 // Remove control character
		}
		if r == 127 { // DEL character
			return -1
		}
		return r
	}, text)
}
