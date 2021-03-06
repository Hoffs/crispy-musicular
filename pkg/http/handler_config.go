package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type configPageData struct {
	User           string
	Config         configPageConfig
	PlaylistConfig configPagePlaylistConfig
}

type configPageConfig struct {
	Interval      uint64
	WorkerCount   uint8
	WorkerTimeout uint32
}

type configPagePlaylistConfig struct {
	IgnoreNotOwned  bool
	IgnoreOwned     bool
	SavedIds        []string
	IgnoredIds      []string
	YoutubeSavedIds []string
}

func (h *httpHandler) configHandler(w http.ResponseWriter, r *http.Request) {
	st, err := h.auth.GetState()
	if err != nil {
		h.renderError(w, "no state found", err)
	}

	d := configPageData{
		User: st.User,
		Config: configPageConfig{
			Interval:      h.config.RunIntervalSeconds,
			WorkerCount:   h.config.WorkerCount,
			WorkerTimeout: h.config.WorkerTimeoutSeconds,
		},
		PlaylistConfig: configPagePlaylistConfig{
			IgnoreNotOwned:  h.config.IgnoreNotOwnedPlaylists,
			IgnoreOwned:     h.config.IgnoreOwnedPlaylists,
			SavedIds:        h.config.SavedPlaylistIds,
			IgnoredIds:      h.config.IgnoredPlaylistIds,
			YoutubeSavedIds: h.config.YoutubeSavedPlaylistIds,
		},
	}
	h.t.renderTemplate(w, "config.tmpl", &d)
}

type configEditPageUserPlaylist struct {
	URI     string
	URIAttr template.HTMLAttr
	Name    string
}
type configEditPageData struct {
	User           string
	Config         configPageConfig
	PlaylistConfig configPagePlaylistConfig
	Playlists      []configEditPageUserPlaylist
}

func (h *httpHandler) editConfigHandler(w http.ResponseWriter, r *http.Request) {
	st, err := h.auth.GetState()
	if err != nil {
		h.renderError(w, "No state found", err)
	}

	c := h.spotAuth.NewClient(&oauth2.Token{RefreshToken: st.RefreshToken})
	p, err := loadUserPlaylists(&c)
	if err != nil {
		h.renderError(w, "Failed to load user playlists", err)
		return
	}

	d := configEditPageData{
		User: st.User,
		Config: configPageConfig{
			Interval:      h.config.RunIntervalSeconds,
			WorkerCount:   h.config.WorkerCount,
			WorkerTimeout: h.config.WorkerTimeoutSeconds,
		},
		PlaylistConfig: configPagePlaylistConfig{
			IgnoreNotOwned:  h.config.IgnoreNotOwnedPlaylists,
			IgnoreOwned:     h.config.IgnoreOwnedPlaylists,
			SavedIds:        h.config.SavedPlaylistIds,
			IgnoredIds:      h.config.IgnoredPlaylistIds,
			YoutubeSavedIds: h.config.YoutubeSavedPlaylistIds,
		},
		Playlists: p,
	}
	h.t.renderTemplate(w, "config_edit.tmpl", &d)
}

func loadUserPlaylists(c *spotify.Client) (p []configEditPageUserPlaylist, err error) {
	limit := 50
	playlists, err := c.CurrentUsersPlaylistsOpt(&spotify.Options{Limit: &limit})
	if err != nil {
		return
	}

	for {
		for _, v := range playlists.Playlists {
			p = append(p, configEditPageUserPlaylist{URI: string(v.URI), URIAttr: template.HTMLAttr("data-uri=" + v.URI), Name: v.Name})
		}

		err = c.NextPage(playlists)
		if err == spotify.ErrNoMorePages {
			err = nil
			return
		}

		if err != nil {
			return
		}
	}
}

func (h *httpHandler) reloadConfigHandler(w http.ResponseWriter, r *http.Request) {
	err := h.config.Reload()
	if err != nil {
		log.Error().Err(err).Msg("handler_config: failed to reload config")
		http.Error(w, "Failed to reload config", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Reloaded")
}

func (h *httpHandler) saveConfigHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("failed to parse form data")
		http.Error(w, "Failed to parse form data", 500)
		return
	}

	interval, err := strconv.ParseUint(r.PostForm.Get("interval"), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse interval")
		http.Error(w, "Incorrect values", 400)
		return
	}

	workers, err := strconv.ParseUint(r.PostForm.Get("workers"), 10, 8)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse workers")
		http.Error(w, "Incorrect values", 400)
		return
	}

	timeout, err := strconv.ParseUint(r.PostForm.Get("timeout"), 10, 32)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse timeout")
		http.Error(w, "Incorrect values", 400)
		return
	}

	ignoreNotOwnedValue := r.PostForm.Get("ignore_not_owned")
	var ignoreNotOwned bool
	if ignoreNotOwnedValue != "" {
		ignoreNotOwned, err = strconv.ParseBool(ignoreNotOwnedValue)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse ignore_not_owned")
			http.Error(w, "Incorrect values", 400)
			return
		}
	}

	ignoreOwnedValue := r.PostForm.Get("ignore_owned")
	var ignoreOwned bool
	if ignoreOwnedValue != "" {
		ignoreOwned, err = strconv.ParseBool(ignoreOwnedValue)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse ignore_owned")
			http.Error(w, "Incorrect values", 400)
			return
		}
	}

	savedIds := parseUriList(r.PostForm.Get("saved"))
	ignoredIds := parseUriList(r.PostForm.Get("ignored"))
	youtubeSavedIds := parseYoutubeList(r.PostForm.Get("youtube_saved"))

	cCopy := (*h.config)
	cCopy.RunIntervalSeconds = interval
	cCopy.WorkerCount = uint8(workers)
	cCopy.WorkerTimeoutSeconds = uint32(timeout)
	cCopy.SavedPlaylistIds = savedIds
	cCopy.IgnoredPlaylistIds = ignoredIds
	cCopy.IgnoreNotOwnedPlaylists = ignoreNotOwned
	cCopy.IgnoreOwnedPlaylists = ignoreOwned
	cCopy.YoutubeSavedPlaylistIds = youtubeSavedIds

	err = h.config.Update(&cCopy)
	if err != nil {
		log.Error().Err(err).Msg("failed to update config")
		http.Error(w, http.StatusText(500), 500)
		return
	}

	http.Redirect(w, r, "/config", http.StatusFound)
}

func parseUriList(in string) (s []string) {
	uris := strings.Fields(in)
	for _, v := range uris {
		id := strings.Replace(v, "spotify:playlist:", "", 1)
		s = append(s, id)
	}

	return
}

func parseYoutubeList(in string) (s []string) {
	uris := strings.Fields(in)
	for _, v := range uris {
		id := strings.Replace(v, "https://www.youtube.com/playlist?list=", "", 1)
		s = append(s, id)
	}

	return
}
