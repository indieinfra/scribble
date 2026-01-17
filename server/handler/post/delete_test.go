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

type stubDeleteStore struct {
	deletedURL     string
	undeleteURL    string
	undeleteNew    bool
	deleteCalled   bool
	undeleteCalled bool
}

func (s *stubDeleteStore) ExistsBySlug(context.Context, string) (bool, error) { return false, nil }
func (s *stubDeleteStore) Create(context.Context, util.Mf2Document) (string, bool, error) {
	return "", false, nil
}
func (s *stubDeleteStore) Update(context.Context, string, map[string][]any, map[string][]any, any) (string, error) {
	return "", nil
}
func (s *stubDeleteStore) Delete(_ context.Context, url string) error {
	s.deleteCalled = true
	s.deletedURL = url
	return nil
}
func (s *stubDeleteStore) Undelete(_ context.Context, url string) (string, bool, error) {
	s.undeleteCalled = true
	if s.undeleteURL == "" {
		s.undeleteURL = url
	}
	return s.undeleteURL, s.undeleteNew, nil
}
func (s *stubDeleteStore) Get(context.Context, string) (*util.Mf2Document, error) { return nil, nil }

func newDeleteState() *state.ScribbleState {
	return &state.ScribbleState{Cfg: &config.Config{Server: config.Server{PublicUrl: "https://example.org"}, Micropub: config.Micropub{MeUrl: "https://example.org"}}}
}

func TestDelete_MissingURL(t *testing.T) {
	st := newDeleteState()
	st.ContentStore = &stubDeleteStore{}
	st.MediaStore = &stubMediaStore{}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	Delete(st, rr, req, map[string]any{}, false)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when url missing, got %d", rr.Code)
	}
}

func TestDelete_Success(t *testing.T) {
	st := newDeleteState()
	store := &stubDeleteStore{}
	st.ContentStore = store
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "delete"}))
	rr := httptest.NewRecorder()

	Delete(st, rr, req, map[string]any{"url": "https://example.org/post"}, false)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if !store.deleteCalled || store.deletedURL != "https://example.org/post" {
		t.Fatalf("expected delete to be called for url, got %+v", store)
	}
}

func TestUndelete_SendsCreatedWhenNewURL(t *testing.T) {
	st := newDeleteState()
	store := &stubDeleteStore{undeleteURL: "https://example.org/new", undeleteNew: true}
	st.ContentStore = store
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "undelete"}))
	rr := httptest.NewRecorder()

	Delete(st, rr, req, map[string]any{"url": "https://example.org/post"}, true)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 for undelete with new url, got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "https://example.org/new" {
		t.Fatalf("expected Location header set")
	}
	if !store.undeleteCalled {
		t.Fatalf("expected undelete to be called")
	}
}

func TestUndelete_NoContentWhenSameURL(t *testing.T) {
	st := newDeleteState()
	store := &stubDeleteStore{undeleteURL: "https://example.org/post", undeleteNew: false}
	st.ContentStore = store
	st.MediaStore = &stubMediaStore{}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "undelete"}))
	rr := httptest.NewRecorder()

	Delete(st, rr, req, map[string]any{"url": "https://example.org/post"}, true)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 when url unchanged, got %d", rr.Code)
	}
}
