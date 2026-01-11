package get

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/indieinfra/scribble/server/resp"
	"github.com/indieinfra/scribble/server/state"
)

type Service struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Photo string `json:"photo"`
}

type SyndicateTo struct {
	Uid     string  `json:"uid"`
	Name    string  `json:"name"`
	Service Service `json:"service"`
}

type Config struct {
	MediaEndpoint string        `json:"media-endpoint"`
	SyndicateTo   []SyndicateTo `json:"syndicate-to"`
}

func HandleConfig(st *state.ScribbleState, w http.ResponseWriter, r *http.Request) {
	cfgOut := Config{
		MediaEndpoint: fmt.Sprintf("%v/media", st.Cfg.Server.PublicUrl),
		SyndicateTo:   []SyndicateTo{},
	}

	err := json.NewEncoder(w).Encode(cfgOut)
	if err != nil {
		resp.WriteInternalServerError(w, "failed to encode configuration data")
		return
	}

	resp.WriteNoContent(w)
}
