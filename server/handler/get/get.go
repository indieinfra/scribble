package get

import (
	"fmt"
	"net/http"

	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/state"
)

func DispatchGet(st *state.ScribbleState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		switch q {
		case "config":
			HandleConfig(st, w, r)
		case "source":
			HandleSource(st, w, r)
		case "syndicate-to":
			HandleSyndicateTo(st, w, r)
		default:
			resp.WriteInvalidRequest(w, fmt.Sprintf("Unknown query: %q", q))
		}
	}
}
