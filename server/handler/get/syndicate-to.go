package get

import (
	"net/http"

	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/state"
)

func HandleSyndicateTo(st *state.ScribbleState, w http.ResponseWriter, r *http.Request) {
	// TODO: Implement syndicate-to retrieval
	resp.WriteOK(w, map[string]any{
		"syndicate-to": []any{},
	})
}
