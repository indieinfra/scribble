package get

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/server/state"
)

func newGetState() *state.ScribbleState {
	return &state.ScribbleState{Cfg: &config.Config{Server: config.Server{PublicUrl: "https://example.org"}}}
}

func TestDispatchGet_UnknownQuery(t *testing.T) {
	st := newGetState()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?q=unknown", nil)

	DispatchGet(st)(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleConfig(t *testing.T) {
	st := newGetState()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?q=config", nil)

	HandleConfig(st, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var cfg Config
	if err := json.Unmarshal(rr.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if cfg.MediaEndpoint != "https://example.org/media" {
		t.Fatalf("unexpected media endpoint %q", cfg.MediaEndpoint)
	}
}

func TestHandleSyndicateTo(t *testing.T) {
	st := newGetState()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?q=syndicate-to", nil)

	HandleSyndicateTo(st, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
