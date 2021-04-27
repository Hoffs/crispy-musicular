package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/rand"
	"github.com/rs/zerolog/log"
)

const authCookieName = "CrispyAuth"

func (h *httpHandler) callbackHandler(w http.ResponseWriter, r *http.Request) {
	tok, err := h.spotAuth.Token(h.spotifyState, r)
	if err != nil {
		h.renderError(w, "Could not get token from request", err)
		return
	}

	if st := r.FormValue("state"); st != h.spotifyState {
		h.renderError(w, fmt.Sprintf("State mismatch: expected %s != %s", h.spotifyState, st), err)
		return
	}

	// If we received auth of different user don't do anything.
	st, err := h.auth.GetState()
	if err != nil {
		h.renderError(w, "Failed to load current state", err)
		return
	}

	client := h.spotAuth.NewClient(tok)
	usr, err := client.CurrentUser()
	if err != nil {
		h.renderError(w, "Failed to reach Spotify API", err)
		return
	}

	if st.IsSet() && usr.ID != st.User {
		// This is technically not an error, but whatever.
		h.renderError(w, "Another user already configured", err)
		return
	}

	if !st.IsSet() {
		err := h.auth.SetState(auth.State{RefreshToken: tok.RefreshToken, User: usr.ID})
		if err != nil {
			h.renderError(w, "Failed to update state", err)
			return
		}
	}

	authT, err := rand.String(24)
	if err != nil {
		h.renderError(w, "Failed to generate auth key", err)
		return
	}

	h.authToken = authT

	http.SetCookie(w, createAuthCookie(h.authToken, time.Now().AddDate(1, 0, 0)))
	h.t.renderTemplate(w, "callback.tmpl", &struct{ User string }{User: usr.ID})
}

func (h *httpHandler) authHandler(w http.ResponseWriter, r *http.Request) {
	s, err := rand.String(16)
	if err != nil {
		h.renderError(w, "Failed to generate random state", err)
		return
	}

	// Have just a single state, because no reason to handle concurrent requests (technically this could just be static).
	h.spotifyState = s

	log.Debug().Msgf("Redirecting to auth with spotify with state '%s'", h.spotifyState)
	http.Redirect(w, r, h.spotAuth.AuthURL(h.spotifyState), http.StatusFound)
	return
}

func (h *httpHandler) deauthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add template
	http.SetCookie(w, createAuthCookie(h.authToken, time.Now().AddDate(0, 0, -1)))
	fmt.Fprintf(w, "Logged out!")
}

func (h *httpHandler) authTestHandler(w http.ResponseWriter, r *http.Request) {
	st, err := h.auth.GetState()
	if err != nil {
		http.Error(w, "No state found", http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "Authenticated! Hello %s", st.User)
}

func (h *httpHandler) authGuard(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(authCookieName)
		if err != nil || c.Value != h.authToken {
			log.Debug().Err(err).Msg("Received request with invalid authorization")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}

func createAuthCookie(v string, exp time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     authCookieName,
		Value:    v,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Expires:  exp,
	}
}
