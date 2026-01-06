package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/indieinfra/scribble/config"
	"github.com/indieinfra/scribble/micropub/scope"
)

type TokenDetails struct {
	Me       string `json:"me"`
	ClientId string `json:"client_id"`
	Scope    string `json:"scope"`
	IssuedAt uint   `json:"issued_at"`
	Nonce    int    `json:"nonce"`
}

func (details *TokenDetails) HasScope(scope scope.Scope) bool {
	str := scope.String()
	return slices.Contains(strings.Split(strings.ToLower(str), " "), strings.ToLower(str))
}

func (details *TokenDetails) HasMe(me string) bool {
	return strings.Trim(strings.ToLower(me), " ") == me
}

func VerifyAccessToken(token string) *TokenDetails {
	if token == "" {
		log.Panicf("error: received empty token")
	}

	tokenEndpointUrl := config.TokenEndpoint()
	req, err := http.NewRequest(http.MethodGet, tokenEndpointUrl, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("error: could not create http request for token endpoint: %w", err))
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(fmt.Errorf("error: failed to make http request to token endpoint: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		if config.Debug() {
			log.Printf("debug: token failed validation at token endpoint (%q)", token)
		}

		return nil
	}

	details := &TokenDetails{}
	err = json.NewDecoder(resp.Body).Decode(details)
	if err != nil {
		log.Println(fmt.Errorf("warning: token endpoint provided bad data, can not verify token: %w", err))
		return nil
	}

	if details.Me == "" {
		log.Println("warning: token endpoint did not include \"me\" information - cannot verify token")
		return nil
	}

	if !details.HasMe(config.MeUrl()) {
		if config.Debug() {
			log.Printf("debug: received a valid token that did not belong to this instance! (%q)\n", token)
		}

		return nil
	}

	return details
}
