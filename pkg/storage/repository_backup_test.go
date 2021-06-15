package storage

import (
	"testing"
	"time"

	bp "github.com/hoffs/crispy-musicular/pkg/backup"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestAddBackup(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)
	require.NoError(t, err)
}

func TestAddPlaylist(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	p := bp.Playlist{SpotifyId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddPlaylist(&b, &p)
	require.NoError(t, err)
}

func TestAddTrack(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	p := bp.Playlist{SpotifyId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddPlaylist(&b, &p)

	tr := bp.Track{SpotifyId: "S", Name: "N", Artist: "Art", Album: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC()}
	err = r.AddTrack(&b, &p, &tr)
	require.NoError(t, err)
}

func TestGetBackupCount(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)
	require.NoError(t, err)

	b = bp.Backup{UserId: "User2", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)
	require.NoError(t, err)

	b = bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)
	require.NoError(t, err)

	count, err := r.GetBackupCount("User")
	require.NoError(t, err)
	require.EqualValues(t, 2, count)
}

func TestGetBackupStatsSpotify(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	p := bp.Playlist{SpotifyId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddPlaylist(&b, &p)

	tr := bp.Track{SpotifyId: "S", Name: "N", Artist: "Art", Album: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC(), PlaylistId: p.Id}
	err = r.AddTrack(&b, &p, &tr)

	sp, st, yp, yt, err := r.GetBackupData(&b)
	require.NoError(t, err)

	require.EqualValues(t, 1, len(*sp))
	require.Equal(t, p, (*sp)[0])
	require.EqualValues(t, 1, len(*st))
	require.Equal(t, tr, (*st)[0])
	require.EqualValues(t, 0, len(*yp))
	require.EqualValues(t, 0, len(*yt))
}

func TestAddYoutubePlaylist(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	p := bp.YoutubePlaylist{YoutubeId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubePlaylist(&b, &p)
	require.NoError(t, err)
}

func TestAddYoutubeTrack(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	p := bp.YoutubePlaylist{YoutubeId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubePlaylist(&b, &p)

	tr := bp.YoutubeTrack{YoutubeId: "S", Name: "N", ChannelTitle: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubeTrack(&b, &p, &tr)
	require.NoError(t, err)
}

func TestGetBackupStatsYoutube(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	p := bp.YoutubePlaylist{YoutubeId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubePlaylist(&b, &p)

	tr := bp.YoutubeTrack{YoutubeId: "S", Name: "N", ChannelTitle: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC(), PlaylistId: p.Id}
	err = r.AddYoutubeTrack(&b, &p, &tr)

	sp, st, yp, yt, err := r.GetBackupData(&b)
	require.NoError(t, err)

	require.EqualValues(t, 0, len(*sp))
	require.EqualValues(t, 0, len(*st))
	require.EqualValues(t, 1, len(*yp))
	require.Equal(t, p, (*yp)[0])
	require.EqualValues(t, 1, len(*yt))
	require.Equal(t, tr, (*yt)[0])
}

func TestGetBackupStats(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	sp := bp.Playlist{SpotifyId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddPlaylist(&b, &sp)

	str := bp.Track{SpotifyId: "S", Name: "N", Artist: "Art", Album: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC(), PlaylistId: sp.Id}
	err = r.AddTrack(&b, &sp, &str)

	p := bp.YoutubePlaylist{YoutubeId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubePlaylist(&b, &p)

	tr := bp.YoutubeTrack{YoutubeId: "S", Name: "N", ChannelTitle: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC(), PlaylistId: p.Id}
	err = r.AddYoutubeTrack(&b, &p, &tr)

	rsp, rst, ryp, ryt, err := r.GetBackupData(&b)
	require.NoError(t, err)

	require.EqualValues(t, 1, len(*rsp))
	require.Equal(t, sp, (*rsp)[0])
	require.EqualValues(t, 1, len(*rst))
	require.Equal(t, str, (*rst)[0])
	require.EqualValues(t, 1, len(*ryp))
	require.Equal(t, p, (*ryp)[0])
	require.EqualValues(t, 1, len(*ryt))
	require.Equal(t, tr, (*ryt)[0])
}

func TestGetBackupPlaylistCount(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	sp := bp.Playlist{SpotifyId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddPlaylist(&b, &sp)

	p := bp.YoutubePlaylist{YoutubeId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubePlaylist(&b, &p)

	p = bp.YoutubePlaylist{YoutubeId: "S2", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubePlaylist(&b, &p)

	count, err := r.GetBackupPlaylistCount(&b)
	require.NoError(t, err)

	require.EqualValues(t, 3, count)
}

func TestGetBackupTrackCount(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b := bp.Backup{UserId: "User", Started: time.Unix(0, 0).UTC()}
	err = r.AddBackup(&b)

	sp := bp.Playlist{SpotifyId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddPlaylist(&b, &sp)

	str := bp.Track{SpotifyId: "S", Name: "N", Artist: "Art", Album: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC(), PlaylistId: sp.Id}
	err = r.AddTrack(&b, &sp, &str)

	str = bp.Track{SpotifyId: "S2", Name: "N", Artist: "Art", Album: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC(), PlaylistId: sp.Id}
	err = r.AddTrack(&b, &sp, &str)

	p := bp.YoutubePlaylist{YoutubeId: "S", Name: "N", Created: time.Unix(0, 0).UTC()}
	err = r.AddYoutubePlaylist(&b, &p)

	tr := bp.YoutubeTrack{YoutubeId: "S", Name: "N", ChannelTitle: "A", AddedAtToPlaylist: "now", Created: time.Unix(0, 0).UTC(), PlaylistId: p.Id}
	err = r.AddYoutubeTrack(&b, &p, &tr)

	count, err := r.GetBackupTrackCount(&b)
	require.NoError(t, err)

	require.EqualValues(t, 3, count)
}

func TestGetLastBackup(t *testing.T) {
	r, err := NewRepository(":memory:")
	require.NoError(t, err)

	b2 := bp.Backup{UserId: "User", Started: time.Unix(200, 0).UTC()}
	err = r.AddBackup(&b2)

	b1 := bp.Backup{UserId: "User", Started: time.Unix(100, 0).UTC()}
	err = r.AddBackup(&b1)

	b3 := bp.Backup{UserId: "User2", Started: time.Unix(500, 0).UTC()}
	err = r.AddBackup(&b3)

	lastBackup, err := r.GetLastBackup("User")
	require.NoError(t, err)

	require.Equal(t, b2, *lastBackup)
}
