package http

import (
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (h *httpHandler) youtubeCallbackHandler(w http.ResponseWriter, r *http.Request) {
	t, err := h.youtubeAuth.Token(r)
	if err != nil {
		h.renderError(w, "Failed to exchange token", err)
		return
	}

	client, err := h.youtubeAuth.NewClient(t)
	if err != nil {
		h.renderError(w, "Failed to create youtube client", err)
		return
	}

	channels, err := client.Channels.List([]string{"snippet"}).Mine(true).Do()
	if err != nil {
		h.renderError(w, "Failed to get user information", err)
		return
	}

	log.Debug().Msgf("%+v\n", channels)

	if t.RefreshToken == "" {
		h.renderError(w, "Refresh token is empty", errors.New("handler_youtube: drive refresh token is empty"))
		return
	}

	st, err := h.auth.GetState()
	if err != nil {
		h.renderError(w, "Failed to get state", err)
		return
	}

	if !st.IsSet() {
		h.renderError(w, "User state is not set", errors.New("handler_youtube: user state is not set"))
		return
	}

	st.YoutubeRefreshToken = t.RefreshToken
	err = h.auth.SetState(st)
	if err != nil {
		h.renderError(w, "Failed to update state", err)
		return
	}

	h.t.renderTemplate(w, "youtube_callback.tmpl", &struct{ User string }{User: channels.Items[0].Snippet.Title})
}

func (h *httpHandler) youtubeAuthHandler(w http.ResponseWriter, r *http.Request) {
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
		h.youtubeAuth.AuthURL(),
		false,
		"",
	}

	if st.YoutubeRefreshToken != "" {
		s, err := h.youtubeAuth.FromRefreshToken(st.YoutubeRefreshToken)
		if err != nil {
			log.Error().Err(err).Msg("handler_youtube: failed to create driver service")
		} else {
			channels, err := s.Channels.List([]string{"snippet"}).Mine(true).Do()
			if err != nil {
				log.Error().Err(err).Msg("handler_youtube: failed to get about drive user")
			} else {
				d.Connected = true
				d.User = channels.Items[0].Snippet.Title
			}
		}
	}

	h.t.renderTemplate(w, "youtube_auth.tmpl", d)
	return
}
