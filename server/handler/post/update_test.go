package post

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/server/auth"
	"github.com/indieinfra/scribble/server/state"
	"github.com/indieinfra/scribble/server/util"
)

type stubUpdateStore struct {
	lastURL      string
	replacements map[string][]any
	additions    map[string][]any
	deletions    any
	newURL       string
}

func (s *stubUpdateStore) ExistsBySlug(context.Context, string) (bool, error) { return false, nil }
func (s *stubUpdateStore) Create(context.Context, util.Mf2Document) (string, bool, error) {
	return "", false, nil
}
func (s *stubUpdateStore) Update(_ context.Context, url string, repl map[string][]any, add map[string][]any, del any) (string, error) {
	s.lastURL = url
	s.replacements = repl
	s.additions = add
	s.deletions = del
	if s.newURL != "" {
		return s.newURL, nil
	}
	return url, nil
}
func (s *stubUpdateStore) Delete(context.Context, string) error { return nil }
func (s *stubUpdateStore) Undelete(context.Context, string) (string, bool, error) {
	return "", false, nil
}
func (s *stubUpdateStore) Get(context.Context, string) (*util.Mf2Document, error) { return nil, nil }

func TestGetDeletionsArray(t *testing.T) {
	input := map[string]any{"delete": []any{"category", "photo"}}
	result, err := getDeletions(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	props, ok := result.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", result)
	}
	if len(props) != 2 || props[0] != "category" || props[1] != "photo" {
		t.Fatalf("unexpected props: %#v", props)
	}
}

func TestGetDeletionsMap(t *testing.T) {
	input := map[string]any{"delete": map[string]any{"category": []any{"foo", "bar"}, "tags": "one"}}
	result, err := getDeletions(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	props, ok := result.(map[string][]any)
	if !ok {
		t.Fatalf("expected map[string][]any, got %T", result)
	}
	if len(props["category"]) != 2 || len(props["tags"]) != 1 {
		t.Fatalf("unexpected values: %#v", props)
	}
}

func TestGetDeletionsInvalid(t *testing.T) {
	input := map[string]any{"delete": 123}
	if _, err := getDeletions(input); err == nil {
		t.Fatalf("expected error for invalid delete type")
	}
}

func TestUpdateRejectsNonJSON(t *testing.T) {
	st := &state.ScribbleState{Cfg: &config.Config{Micropub: config.Micropub{MeUrl: "https://example.org"}}}
	st.ContentStore = &stubUpdateStore{}
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "update"}))
	rr := httptest.NewRecorder()

	Update(st, rr, req, map[string]any{})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-json, got %d", rr.Code)
	}
}

func TestUpdateMissingURL(t *testing.T) {
	st := &state.ScribbleState{Cfg: &config.Config{Micropub: config.Micropub{MeUrl: "https://example.org"}}}
	st.ContentStore = &stubUpdateStore{}
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "update"}))
	rr := httptest.NewRecorder()

	Update(st, rr, req, map[string]any{})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing url, got %d", rr.Code)
	}
}

func TestUpdateInsufficientScope(t *testing.T) {
	st := &state.ScribbleState{Cfg: &config.Config{Micropub: config.Micropub{MeUrl: "https://example.org"}}}
	st.ContentStore = &stubUpdateStore{}
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	// No token provided
	rr := httptest.NewRecorder()

	Update(st, rr, req, map[string]any{"url": "https://example.org/post"})

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing scope, got %d", rr.Code)
	}
}

func TestUpdateWritesCreatedWhenURLChanges(t *testing.T) {
	st := &state.ScribbleState{Cfg: &config.Config{Micropub: config.Micropub{MeUrl: "https://example.org"}}}
	store := &stubUpdateStore{newURL: "https://example.org/new"}
	st.ContentStore = store
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "update"}))
	rr := httptest.NewRecorder()

	replacements := map[string]any{"name": []any{"New"}}
	additions := map[string]any{"category": "go"}
	deletions := map[string]any{"category": []any{"old"}}

	Update(st, rr, req, map[string]any{
		"url":     "https://example.org/post",
		"replace": replacements,
		"add":     additions,
		"delete":  deletions,
	})

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 when url changes, got %d", rr.Code)
	}
	if store.lastURL != "https://example.org/post" {
		t.Fatalf("unexpected url sent to store: %q", store.lastURL)
	}
}

func TestUpdateWritesNoContentWhenURLSame(t *testing.T) {
	st := &state.ScribbleState{Cfg: &config.Config{Micropub: config.Micropub{MeUrl: "https://example.org"}}}
	store := &stubUpdateStore{}
	st.ContentStore = store
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "update"}))
	rr := httptest.NewRecorder()

	Update(st, rr, req, map[string]any{"url": "https://example.org/post"})

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 when url unchanged, got %d", rr.Code)
	}
}
