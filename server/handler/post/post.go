package post

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/state"
)

func DispatchPost(st *state.ScribbleState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := ReadBody(st.Cfg, w, r)
		if body == nil {
			return
		}

		actionRaw, ok := body["action"]
		if !ok {
			actionRaw = "create"
		}

		action, ok := actionRaw.(string)
		if !ok {
			resp.WriteInvalidRequest(w, fmt.Sprintf("Action must be a string, got %v", action))
			return
		}

		delete(body, "action")

		switch strings.ToLower(action) {
		case "create":
			Create(st, w, r, body)
		case "update":
			Update(st, w, r, body)
		case "delete":
			Delete(st, w, r, body, false)
		case "undelete":
			Delete(st, w, r, body, true)
		default:
			resp.WriteInvalidRequest(w, fmt.Sprintf("Unknown action: %q", action))
		}
	}
}
