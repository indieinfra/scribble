package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/micropub/resp"
	"github.com/indieinfra/scribble/micropub/scope"
)

type tokenKeyType struct{}

var tokenKey = tokenKeyType{}

// function ValidateTokenMiddleware wraps a downstream handler. At execution time,
// it extracts a Bearer token from the Authorization header, if any. If the Authorization
// header is not present, or does not contain a Bearer token, it aborts the request.
// If the token is present, it performs the VerifyAccessToken routine which makes a downstream
// http request to the defined token endpoint to validate the token.
func ValidateTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		scheme, token, ok := strings.Cut(authorization, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") {
			resp.WriteHttpError(w, http.StatusBadRequest, "A bearer token is required. Please ensure Authorization header is in Bearer format")
			return
		}

		token = strings.TrimSpace(token)
		if token == "" {
			resp.WriteHttpError(w, http.StatusBadRequest, "Received bearer token is empty")
			return
		}

		details := VerifyAccessToken(token)
		if details == nil {
			resp.WriteHttpError(w, http.StatusUnauthorized, "Token validation failed. Please try again with a valid token.")
			return
		}

		ctx := context.WithValue(r.Context(), tokenKey, details)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// function RequireScopeMiddleware wraps a downstream handler. At execution time,
// the middleware expects a valid token to be available in the request context.
// The middleware will access the stored token details and validate the token
// contains the required scopes. Without the required scopes, the middleware will
// abort the request.
func RequireScopeMiddleware(scopes []scope.Scope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			details, ok := r.Context().Value(tokenKey).(*TokenDetails)
			if !ok {
				resp.WriteHttpError(w, http.StatusUnauthorized, "Request is missing token")
				return
			}

			for _, scope := range scopes {
				if !details.HasScope(scope) {
					if config.Debug() {
						log.Printf("debug: received a valid token, but failed scope check (want %v, have %q)", scope, details.Scope)
					}

					resp.WriteHttpError(w, http.StatusForbidden, "Insufficient scope for request.")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
