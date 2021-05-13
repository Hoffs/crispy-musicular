package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/backup"
	"github.com/hoffs/crispy-musicular/pkg/config"
	"github.com/hoffs/crispy-musicular/pkg/drive"
	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
)

func RegisterHandlers(c *config.AppConfig, auth auth.Service, b backup.Service) error {
	h := &httpHandler{
		auth:      auth,
		spotAuth:  spotify.NewAuthenticator(c.SpotifyCallback, spotify.ScopePlaylistReadPrivate),
		driveAuth: drive.NewAuthenticator(c.DriveId, c.DriveSecret, c.DriveCallback),
		backuper:  b,
		config:    c,
		t:         NewTemplater("templates", os.Getenv("DEBUG") == ""),
	}

	http.HandleFunc("/auth", methodGuard(http.MethodGet, h.authHandler))
	http.HandleFunc("/callback", methodGuard(http.MethodGet, h.callbackHandler))
	http.HandleFunc("/deauth", methodGuard(http.MethodGet, h.deauthHandler))

	http.HandleFunc("/auth_test", methodGuard(http.MethodGet, debugGuard(h.authGuard(h.authTestHandler))))

	http.HandleFunc("/home", methodGuard(http.MethodGet, h.authGuard(h.homeHandler)))
	http.HandleFunc("/backup/start", methodGuard(http.MethodPost, h.authGuard(h.backupStartHandler)))

	http.HandleFunc("/config", methodGuard(http.MethodGet, h.authGuard(h.configHandler)))
	http.HandleFunc("/config/edit", methodGuard(http.MethodGet, h.authGuard(h.editConfigHandler)))
	http.HandleFunc("/config/edit/save", methodGuard(http.MethodPost, h.authGuard(h.saveConfigHandler)))
	http.HandleFunc("/config/reload", methodGuard(http.MethodPost, h.authGuard(h.reloadConfigHandler)))

	http.HandleFunc("/drive/auth", methodGuard(http.MethodGet, h.authGuard(h.driveAuthHandler)))
	http.HandleFunc("/drive/callback", methodGuard(http.MethodGet, h.authGuard(h.driveCallbackHandler)))

	return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
}

type httpHandler struct {
	auth         auth.Service
	spotAuth     spotify.Authenticator
	driveAuth    drive.Authenticator
	backuper     backup.Service
	config       *config.AppConfig
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

// Pages that only work in debug env
func debugGuard(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEBUG") == "" {
			http.NotFound(w, r)
			return
		}

		handler(w, r)
	}
}

// Doens't really work well if same route should handle GET/POST
func methodGuard(method string, handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			log.Debug().Msgf("http: received request at path %s with method %s, expected: %s", r.URL.Path, r.Method, method)
			http.NotFound(w, r)
			return
		}

		handler(w, r)
	}
}
