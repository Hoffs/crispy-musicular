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
`

func TestLoadConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "testconf")
	require.NoError(t, err)

	defer os.Remove(f.Name())
	ioutil.WriteFile(f.Name(), []byte(config_file), fs.ModeAppend)

	config, err := Load(f.Name())

	require.NoError(t, err)
	require.Equal(t, config.RunIntervalSeconds, uint64(360))
}

var config_file_invalid = `
runIntervalSeconds: 260
`

func TestLoadConfigInvalidValues(t *testing.T) {
	f, err := ioutil.TempFile("", "testconf")
	require.NoError(t, err)

	defer os.Remove(f.Name())
	ioutil.WriteFile(f.Name(), []byte(config_file_invalid), fs.ModeAppend)

	_, err = Load(f.Name())

	require.Error(t, err)
}
