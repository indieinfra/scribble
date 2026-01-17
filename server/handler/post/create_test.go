package post

import (
	"context"
	"testing"

	"github.com/indieinfra/scribble/server/util"
)

type stubStore struct{ exists bool }

func (s *stubStore) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return s.exists, nil
}
func (s *stubStore) Create(context.Context, util.Mf2Document) (string, bool, error) {
	return "", false, nil
}
func (s *stubStore) Update(context.Context, string, map[string][]any, map[string][]any, any) (string, error) {
	return "", nil
}
func (s *stubStore) Delete(context.Context, string) error                   { return nil }
func (s *stubStore) Undelete(context.Context, string) (string, bool, error) { return "", false, nil }
func (s *stubStore) Get(context.Context, string) (*util.Mf2Document, error) { return nil, nil }

func TestDeriveSuggestedSlug(t *testing.T) {
	t.Run("mp-slug wins", func(t *testing.T) {
		doc := util.Mf2Document{Properties: map[string][]any{"mp-slug": {"custom"}}}
		if got := deriveSuggestedSlug(&doc); got != "custom" {
			t.Fatalf("expected mp-slug, got %q", got)
		}
	})

	t.Run("generated slug", func(t *testing.T) {
		doc := util.Mf2Document{Properties: map[string][]any{"name": {"Hello"}}}
		if got := deriveSuggestedSlug(&doc); got != "hello" {
			t.Fatalf("expected generated slug, got %q", got)
		}
	})

	t.Run("uuid fallback", func(t *testing.T) {
		doc := util.Mf2Document{Properties: map[string][]any{"photo": {"noop"}}}
		got := deriveSuggestedSlug(&doc)
		if got == "" {
			t.Fatalf("expected uuid fallback slug")
		}
	})
}

func TestEnsureUniqueSlug(t *testing.T) {
	t.Run("returns slug when unique", func(t *testing.T) {
		store := &stubStore{exists: false}
		got, err := ensureUniqueSlug(context.Background(), store, "slug")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "slug" {
			t.Fatalf("expected slug unchanged, got %q", got)
		}
	})

	t.Run("adds suffix when exists", func(t *testing.T) {
		store := &stubStore{exists: true}
		got, err := ensureUniqueSlug(context.Background(), store, "slug")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == "slug" {
			t.Fatalf("expected slug to change when collision")
		}
		if len(got) <= len("slug") {
			t.Fatalf("expected suffix added, got %q", got)
		}
	})
}
