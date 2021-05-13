package http

import (
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (h *httpHandler) driveCallbackHandler(w http.ResponseWriter, r *http.Request) {
	t, err := h.driveAuth.Token(r)
	if err != nil {
		h.renderError(w, "Failed to exchange token", err)
		return
	}

	client, err := h.driveAuth.NewClient(t)
	if err != nil {
		h.renderError(w, "Failed to create drive client", err)
		return
	}

	about, err := client.About.Get().Fields("user/displayName").Do()
	if err != nil {
		h.renderError(w, "Failed to get user information", err)
		return
	}

	files, err := client.Files.List().Q("mimeType = 'application/vnd.google-apps.folder'").Fields("files/id,files/name").Do()
	if err != nil {
		h.renderError(w, "Failed to get user information", err)
		return
	}

	for _, x := range files.Files {
		log.Debug().Msgf("found id %s name %s", x.Id, x.Name)
	}

	if t.RefreshToken == "" {
		h.renderError(w, "Refresh token is empty", errors.New("handler_drive: drive refresh token is empty"))
		return
	}

	st, err := h.auth.GetState()
	if err != nil {
		h.renderError(w, "Failed to get state", err)
		return
	}

	if !st.IsSet() {
		h.renderError(w, "User state is not set", errors.New("handler_drive: user state is not set"))
		return
	}

	st.DriveRefreshToken = t.RefreshToken
	err = h.auth.SetState(st)
	if err != nil {
		h.renderError(w, "Failed to update state", err)
		return
	}

	h.t.renderTemplate(w, "drive_callback.tmpl", &struct{ User string }{User: about.User.DisplayName})
}

func (h *httpHandler) driveAuthHandler(w http.ResponseWriter, r *http.Request) {
	st, err := h.auth.GetState()
	if err != nil {
		h.renderError(w, "Failed to get state", err)
		return
	}

	d := &struct {
		AuthUrl   string
		Connected bool
		User      string
	}{
		h.driveAuth.AuthURL(),
		false,
		"",
	}

	if st.DriveRefreshToken != "" {
		s, err := h.driveAuth.FromRefreshToken(st.DriveRefreshToken)
		if err != nil {
			log.Error().Err(err).Msg("handler_drive: failed to create driver service")
		} else {
			about, err := s.About.Get().Fields("user/displayName").Do()
			if err != nil {
				log.Error().Err(err).Msg("handler_drive: failed to get about drive user")
			} else {
				d.Connected = true
				d.User = about.User.DisplayName
			}
		}
	}

	h.t.renderTemplate(w, "drive_auth.tmpl", d)
	return
}
