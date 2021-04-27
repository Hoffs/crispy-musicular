package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/rand"
)

const authCookieName = "CrispyAuth"

func (h *httpHandler) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add some logging
	tok, err := h.spotAuth.Token(h.spotifyState, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		return
	}

	if st := r.FormValue("state"); st != h.spotifyState {
		http.NotFound(w, r)
		return
		// log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// If we received auth of different user don't do anything.
	st, err := h.auth.GetState()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Failed to load current state", http.StatusInternalServerError)
		return
	}

	client := h.spotAuth.NewClient(tok)
	usr, err := client.CurrentUser()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if st.IsSet() && usr.ID != st.User {
		http.Error(w, "Another user already configured", http.StatusConflict)
		return
	}

	if !st.IsSet() {
		err := h.auth.SetState(auth.State{RefreshToken: tok.RefreshToken, User: usr.ID})
		if err != nil {
			http.Error(w, "Failed to configure user", http.StatusInternalServerError)
			return
		}
	}

	authT, err := rand.String(24)
	if err != nil {
		http.Error(w, "Failed to generate auth key", http.StatusInternalServerError)
		return
	}

	h.authToken = authT

	w.Header().Set("Content-Type", "text/html")
	http.SetCookie(w, createAuthCookie(h.authToken, time.Now().AddDate(1, 0, 0)))
	fmt.Fprintf(w, "Login Completed! Hi %s", usr.ID)
}

func (h *httpHandler) authHandler(w http.ResponseWriter, r *http.Request) {
	s, err := rand.String(16)
	if err != nil {
		http.Error(w, "Failed to generate random state", http.StatusInternalServerError)
		return
	}

	// Have just a single state, because no reason to handle concurrent requests (technically this could just be static).
	h.spotifyState = s

	http.Redirect(w, r, h.spotAuth.AuthURL(h.spotifyState), http.StatusFound)
	return
}

func (h *httpHandler) deauthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.SetCookie(w, createAuthCookie(h.authToken, time.Now().AddDate(0, 0, -1)))
	fmt.Fprintf(w, "Logged out!")
}

func (h *httpHandler) authTestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
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
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: Log
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
