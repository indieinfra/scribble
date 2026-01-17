package util

import "testing"

func TestGenerateSlug(t *testing.T) {
	t.Run("uses name when present", func(t *testing.T) {
		doc := Mf2Document{Properties: map[string][]any{"name": {"Hello World"}}}
		if slug := GenerateSlug(doc); slug != "hello-world" {
			t.Fatalf("expected slug from name, got %q", slug)
		}
	})

	t.Run("falls back to content", func(t *testing.T) {
		doc := Mf2Document{Properties: map[string][]any{"content": {"An interesting post"}}}
		if slug := GenerateSlug(doc); slug != "an-interesting-post" {
			t.Fatalf("expected slug from content, got %q", slug)
		}
	})

	t.Run("combines name and content when name short", func(t *testing.T) {
		doc := Mf2Document{Properties: map[string][]any{"name": {"Hello"}, "content": {"world from scribble today"}}}
		slug := GenerateSlug(doc)
		if slug != "hello-world-from-scribble-today" {
			t.Fatalf("unexpected slug: %q", slug)
		}
	})

	t.Run("empty when no usable fields", func(t *testing.T) {
		doc := Mf2Document{Properties: map[string][]any{"photo": {"http://example.com/img.jpg"}}}
		if slug := GenerateSlug(doc); slug != "" {
			t.Fatalf("expected empty slug when no name/content, got %q", slug)
		}
	})
}
