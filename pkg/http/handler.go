package http

import (
	"fmt"
	"net/http"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/zmb3/spotify"
)

func RegisterHandlers(c *config.AppConfig, auth auth.Service) error {
	h := &httpHandler{
		auth:     auth,
		spotAuth: spotify.NewAuthenticator(c.SpotifyCallback, spotify.ScopePlaylistReadPrivate),
	}

	// Can will the handler access it's state of httpHandler?
	http.HandleFunc("/callback", h.callbackHandler)
	http.HandleFunc("/auth", h.authHandler)
	http.HandleFunc("/deauth", h.deauthHandler)
	http.HandleFunc("/auth_test", h.authGuard(h.authTestHandler))

	return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
}

// CreateHandler
// LoadTemplates in handler at once

type httpHandler struct {
	auth         auth.Service
	spotAuth     spotify.Authenticator
	spotifyState string
	authToken    string
}
