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
