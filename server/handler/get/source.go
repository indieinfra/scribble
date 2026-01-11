package get

import (
	"net/http"

	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/state"
)

func HandleSource(st *state.ScribbleState, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	url := q.Get("url")
	if url == "" {
		resp.WriteInvalidRequest(w, "source requires a url")
		return
	}

	resp.WriteNoContent(w)
}
