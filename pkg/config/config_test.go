package config

import (
	"io/fs"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var config_file = `
runIntervalSeconds: 360
port: 1337
spotifyCallback: http://localhost:1337
workerCount: 12
workerTimeoutSeconds: 500
savedPlaylistIds:
- 3YWsEVozX85ZkwO0d2u8Xx
ignoredPlaylistIds:
- 12345
- 678
youtubeSavedPlaylistIds:
- 1
`

func TestLoadConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "testconf")
	require.NoError(t, err)

	defer os.Remove(f.Name())
	ioutil.WriteFile(f.Name(), []byte(config_file), fs.ModeAppend)

	os.Setenv("SPOTIFY_ID", "AA")
	os.Setenv("SPOTIFY_SECRET", "BB")
	config, err := Load(f.Name())

	require.NoError(t, err)
	require.Equal(t, config.RunIntervalSeconds, uint64(360))
}

var config_file_invalid = `
runIntervalSeconds: 0
port: 1337
spotifyCallback: http://localhost:1337
workerCount: 12
workerTimeoutSeconds: 500
`

func TestLoadConfigInvalidValues(t *testing.T) {
	f, err := ioutil.TempFile("", "testconf")
	require.NoError(t, err)

	defer os.Remove(f.Name())
	ioutil.WriteFile(f.Name(), []byte(config_file_invalid), fs.ModeAppend)

	os.Setenv("SPOTIFY_ID", "AA")
	os.Setenv("SPOTIFY_SECRET", "BB")
	_, err = Load(f.Name())

	require.Error(t, err)
	require.Contains(t, err.Error(), "appconfig: RunIntervalSeconds must be configured and more than 0")
}

var config_file_updated = `
runIntervalSeconds: 50
port: 2337
spotifyCallback: https://localhost:1337
workerCount: 8
workerTimeoutSeconds: 500
savedPlaylistIds:
- 12
- 24
ignoredPlaylistIds:
- xy
youtubeSavedPlaylistIds:
- 2
- 3
`

func TestReloadConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "testconf")
	require.NoError(t, err)

	defer os.Remove(f.Name())
	ioutil.WriteFile(f.Name(), []byte(config_file), fs.ModeAppend)

	os.Setenv("SPOTIFY_ID", "AA")
	os.Setenv("SPOTIFY_SECRET", "BB")
	config, err := Load(f.Name())

	require.NoError(t, err)
	require.Equal(t, uint64(360), config.RunIntervalSeconds)

	ioutil.WriteFile(f.Name(), []byte(config_file_updated), fs.ModeAppend)
	err = config.Reload()
	require.NoError(t, err)

	require.Equal(t, uint64(50), config.RunIntervalSeconds)
	require.Equal(t, "http://localhost:1337", config.SpotifyCallback)
	require.Equal(t, uint32(1337), config.Port)
	require.Equal(t, uint8(8), config.WorkerCount)

	require.Equal(t, "12", config.SavedPlaylistIds[0])
	require.Equal(t, "24", config.SavedPlaylistIds[1])
	require.Equal(t, 2, len(config.SavedPlaylistIds))
	require.Equal(t, "xy", config.IgnoredPlaylistIds[0])
	require.Equal(t, 1, len(config.IgnoredPlaylistIds))
	require.Equal(t, 2, len(config.YoutubeSavedPlaylistIds))
	require.Equal(t, "2", config.YoutubeSavedPlaylistIds[0])
	require.Equal(t, "3", config.YoutubeSavedPlaylistIds[1])
}

var config_file_updated_invalid = `
runIntervalSeconds: 0
port: 2337
spotifyCallback: https://localhost:1337
workerCount: 8
workerTimeoutSeconds: 500
savedPlaylistIds:
- 12
- 24
ignoredPlaylistIds:
- xy
`

func TestReloadConfigInvalid(t *testing.T) {
	f, err := ioutil.TempFile("", "testconf")
	require.NoError(t, err)

	defer os.Remove(f.Name())
	ioutil.WriteFile(f.Name(), []byte(config_file), fs.ModeAppend)

	os.Setenv("SPOTIFY_ID", "AA")
	os.Setenv("SPOTIFY_SECRET", "BB")
	config, err := Load(f.Name())

	require.NoError(t, err)
	require.Equal(t, uint64(360), config.RunIntervalSeconds)

	ioutil.WriteFile(f.Name(), []byte(config_file_updated_invalid), fs.ModeAppend)
	err = config.Reload()
	require.Error(t, err)

	require.Equal(t, uint64(360), config.RunIntervalSeconds)
	require.Equal(t, "http://localhost:1337", config.SpotifyCallback)
	require.Equal(t, uint32(1337), config.Port)
	require.Equal(t, uint8(12), config.WorkerCount)

	require.Equal(t, "3YWsEVozX85ZkwO0d2u8Xx", config.SavedPlaylistIds[0])
	require.Equal(t, 1, len(config.SavedPlaylistIds))
	require.Equal(t, "12345", config.IgnoredPlaylistIds[0])
	require.Equal(t, "678", config.IgnoredPlaylistIds[1])
	require.Equal(t, 2, len(config.IgnoredPlaylistIds))
	require.Equal(t, 1, len(config.YoutubeSavedPlaylistIds))
}
