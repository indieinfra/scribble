package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/server/handler/get"
	"github.com/indieinfra/scribble/server/handler/post"
	"github.com/indieinfra/scribble/server/handler/upload"
	"github.com/indieinfra/scribble/server/middleware"
	"github.com/indieinfra/scribble/server/state"
	"github.com/indieinfra/scribble/storage/content"
	"github.com/indieinfra/scribble/storage/media"
)

func StartServer(cfg *config.Config) {
	log.Println("initializing...")
	state, err := initialize(&state.ScribbleState{Cfg: cfg})
	if err != nil {
		log.Fatalf("initialization failed: %v", err)
		return
	}

	// Setup cleanup on shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("shutting down...")
		cleanup(state)
		os.Exit(0)
	}()

	log.Println("configuring routes...")
	mux := http.NewServeMux()
	mux.Handle("GET /", middleware.ValidateTokenMiddleware(state.Cfg, get.DispatchGet(state)))
	mux.Handle("POST /", middleware.ValidateTokenMiddleware(state.Cfg, post.DispatchPost(state)))
	mux.Handle("POST /media", middleware.ValidateTokenMiddleware(state.Cfg, upload.HandleMediaUpload(state)))

	bindAddress := fmt.Sprintf("%v:%v", state.Cfg.Server.Address, state.Cfg.Server.Port)
	log.Printf("serving http requests on %q", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, mux))
}

func initialize(state *state.ScribbleState) (*state.ScribbleState, error) {
	contentStore, err := initializeContentStore(&state.Cfg.Content)
	if err != nil {
		return nil, err
	}
	state.ContentStore = contentStore
	state.MediaStore = &media.NoopMediaStore{}

	return state, nil
}

func initializeContentStore(cfg *config.Content) (content.ContentStore, error) {
	if cfg.Strategy == "git" {
		store, err := content.NewGitContentStore(cfg.Git)
		if err != nil {
			return nil, err
		}

		return store, nil
	}

	return nil, fmt.Errorf("...unknown content strategy %q", cfg.Strategy)
}

func cleanup(state *state.ScribbleState) {
	// Cleanup git content store if applicable
	if gitStore, ok := state.ContentStore.(*content.GitContentStore); ok {
		if err := gitStore.Cleanup(); err != nil {
			log.Printf("error during cleanup: %v", err)
		}
	}
}
