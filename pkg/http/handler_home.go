package http

import (
	"net/http"
	"time"
)

type homePageData struct {
	User           string
	Config         homePageConfig
	PlaylistConfig homePagePlaylistConfig
	Stats          homePageStats
}

type homePageConfig struct {
	Interval      uint64
	WorkerCount   uint8
	WorkerTimeout uint32
}

type homePagePlaylistConfig struct {
	IgnoreNotOwned bool
	SavedIds       []string
	IgnoredIds     []string
}

type homePageStats struct {
	LastStartedAt  formattedTime
	LastFinishedAt formattedTime
	LastPlaylists  int64
	LastTracks     int64
	TotalBackups   int64
}

type formattedTime struct {
	time time.Time
}

func (ft formattedTime) String() string {
	if ft.time == (time.Time{}) {
		return ""
	}
	return ft.time.Format("2006-01-02 15:04:05 -0700")
}

func (h *httpHandler) homeHandler(w http.ResponseWriter, r *http.Request) {
	st, err := h.auth.GetState()
	// TODO: figure out something to handle these errors easier
	if err != nil {
		h.renderError(w, "No state found", err)
	}

	backupStats, err := h.backuper.GetBackupStats(st.User)
	if err != nil {
		h.renderError(w, "Could not get last backup stats", err)
		return
	}

	d := homePageData{
		User: st.User,
		Config: homePageConfig{
			Interval:      h.config.RunIntervalSeconds,
			WorkerCount:   h.config.WorkerCount,
			WorkerTimeout: h.config.WorkerTimeoutSeconds,
		},
		PlaylistConfig: homePagePlaylistConfig{
			IgnoreNotOwned: h.config.IgnoreNotOwnedPlaylists,
			SavedIds:       h.config.SavedPlaylistIds,
			IgnoredIds:     h.config.IgnoredPlaylistIds,
		},
		Stats: homePageStats{
			LastStartedAt:  formattedTime{backupStats.StartedAt},
			LastFinishedAt: formattedTime{backupStats.FinishedAt},
			LastPlaylists:  backupStats.PlaylistCount,
			LastTracks:     backupStats.TrackCount,
			TotalBackups:   backupStats.TotalBackups,
		},
	}
	h.t.renderTemplate(w, "home.tmpl", &d)
}
