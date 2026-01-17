package upload

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/server/auth"
	"github.com/indieinfra/scribble/server/state"
)

type fakeMediaStore struct{ called bool }

func (f *fakeMediaStore) Upload(_ context.Context, _ *multipart.File, _ *multipart.FileHeader) (string, error) {
	f.called = true
	return "https://media.example.org/file", nil
}
func (f *fakeMediaStore) Delete(context.Context, string) error { return nil }

func newUploadState() *state.ScribbleState {
	return &state.ScribbleState{Cfg: &config.Config{Server: config.Server{Limits: config.ServerLimits{MaxPayloadSize: 2_000_000, MaxFileSize: 1_000_000, MaxMultipartMem: 2_000_000}}, Micropub: config.Micropub{MeUrl: "https://example.org"}}}
}

func TestHandleMediaUpload_TokenConflict(t *testing.T) {
	st := newUploadState()
	ms := &fakeMediaStore{}
	st.MediaStore = ms

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("access_token", "bodytoken")
	fw, _ := w.CreateFormFile("file", "a.txt")
	fw.Write([]byte("hello"))
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/media", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "media"}))

	rr := httptest.NewRecorder()
	HandleMediaUpload(st)(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for token conflict, got %d", rr.Code)
	}
	if ms.called {
		t.Fatalf("media store should not be called on token conflict")
	}
}

func TestHandleMediaUpload_Success(t *testing.T) {
	st := newUploadState()
	ms := &fakeMediaStore{}
	st.MediaStore = ms

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "a.txt")
	fw.Write([]byte("hello"))
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/media", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req = req.WithContext(auth.AddToken(req.Context(), &auth.TokenDetails{Me: st.Cfg.Micropub.MeUrl, Scope: "media"}))

	rr := httptest.NewRecorder()
	HandleMediaUpload(st)(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	if !ms.called {
		t.Fatalf("expected media store upload to be called")
	}
}
