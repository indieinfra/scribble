package server

import (
	"log"
	"net/http"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/micropub/auth"
	"github.com/indieinfra/scribble/micropub/scope"
)

func StartServer() {
	mux := http.NewServeMux()
	// mux.Handle("POST /create", scoped(handler.HandleCreate, scope.Create))

	bindAddress := config.BindAddress()
	log.Printf("serving http requests on %q", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, mux))
}

func scoped(h http.HandlerFunc, scopes ...scope.Scope) http.Handler {
	return chain(
		h,
		auth.ValidateTokenMiddleware,
		auth.RequireScopeMiddleware(scopes),
	)
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares); i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}
