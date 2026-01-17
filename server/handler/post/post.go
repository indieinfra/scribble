package post

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/indieinfra/scribble/server/auth"
	"github.com/indieinfra/scribble/server/middleware"
	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/state"
)

func DispatchPost(st *state.ScribbleState) http.HandlerFunc {
	handlers := map[string]func(*state.ScribbleState, http.ResponseWriter, *http.Request, *ParsedBody){
		"create": Create,
		"update": func(st *state.ScribbleState, w http.ResponseWriter, r *http.Request, body *ParsedBody) {
			Update(st, w, r, body.Data)
		},
		"delete": func(st *state.ScribbleState, w http.ResponseWriter, r *http.Request, body *ParsedBody) {
			Delete(st, w, r, body.Data, false)
		},
		"undelete": func(st *state.ScribbleState, w http.ResponseWriter, r *http.Request, body *ParsedBody) {
			Delete(st, w, r, body.Data, true)
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		parsed, ok := ReadBody(st.Cfg, w, r)
		if !ok {
			return
		}
		r, ok = middleware.EnsureTokenForRequest(st.Cfg, w, r, parsed.AccessToken)
		if !ok {
			return
		}
		if parsed.File != nil && parsed.File.File != nil {
			defer parsed.File.File.Close()
		}

		actionRaw, ok := parsed.Data["action"]
		if !ok {
			actionRaw = "create"
		}

		action, ok := actionRaw.(string)
		if !ok {
			resp.WriteInvalidRequest(w, fmt.Sprintf("Action must be a string, got %v", action))
			return
		}

		delete(parsed.Data, "action")

		if handler, ok := handlers[strings.ToLower(action)]; ok {
			handler(st, w, r, parsed)
			return
		}

		resp.WriteInvalidRequest(w, fmt.Sprintf("Unknown action: %q", action))
	}
}

func requireScope(w http.ResponseWriter, r *http.Request, scope auth.Scope) bool {
	if !auth.RequestHasScope(r, scope) {
		resp.WriteInsufficientScope(w, fmt.Sprintf("no %s scope", scope.String()))
		return false
	}
	return true
}
