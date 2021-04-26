package http

import (
	"fmt"
	"net/http"

	"github.com/hoffs/crispy-musicular/pkg/auth"
	"github.com/hoffs/crispy-musicular/pkg/rand"
)

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

	// TODO: Somehow authorize requester as the current configured user (set special token as cookie/session?)
	// so that he could de-auth or see whatever
	w.Header().Set("Content-Type", "text/html")
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
