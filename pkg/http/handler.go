package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
)

func RegisterHandlers(c *config.AppConfig, auth auth.Service) error {
	h := &httpHandler{
		auth:     auth,
		spotAuth: spotify.NewAuthenticator(c.SpotifyCallback, spotify.ScopePlaylistReadPrivate),
		t:        NewTemplater("templates", os.Getenv("DEBUG") == ""),
	}

	// Can will the handler access it's state of httpHandler?
	http.HandleFunc("/home", h.authGuard(h.homeHandler))
	http.HandleFunc("/callback", h.callbackHandler)
	http.HandleFunc("/auth", h.authHandler)
	http.HandleFunc("/deauth", h.deauthHandler)
	http.HandleFunc("/auth_test", debugGuard(h.authGuard(h.authTestHandler)))

	return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
}

type httpHandler struct {
	auth         auth.Service
	spotAuth     spotify.Authenticator
	t            *templater
	spotifyState string
	authToken    string
}

func (h *httpHandler) renderError(w http.ResponseWriter, title string, err error) {
	log.Error().Err(err).Msg(title)

	d := &struct {
		Title string
		Text  string
	}{
		Title: title,
		Text:  err.Error(),
	}

	// Don't return error text in non-debug mode.
	if os.Getenv("DEBUG") == "" {
		d.Title = "Unexpected server error"
		d.Text = "Please try again"
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(500)
	h.t.renderTemplate(w, "error.tmpl", d)
}

func debugGuard(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	// Pages that only work in debug env
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEBUG") == "" {
			http.NotFound(w, r)
			return
		}

		handler(w, r)
	}
}
