package config

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

// If saved playlist ids is not empty, "IgnoredPlaylistIds" is not used.
type AppConfig struct {
	RunIntervalSeconds      uint64   `yaml:"runIntervalSeconds"`
	Port                    uint32   `yaml:"port"`
	SpotifyCallback         string   `yaml:"spotifyCallback"`
	WorkerCount             uint8    `yaml:"workerCount"`
	WorkerTimeoutSeconds    uint32   `yaml:"workerTimeoutSeconds"`
	SavedPlaylistIds        []string `yaml:"savedPlaylistIds"`
	IgnoredPlaylistIds      []string `yaml:"ignoredPlaylistIds"`
	IgnoreNotOwnedPlaylists bool     `yaml:"ignoreNotOwnedPlaylists"`
	IgnoreOwnedPlaylists    bool     `yaml:"ignoreOwnedPlaylists"`
	JsonDir                 string   `yaml:"jsonDir"`
	DbPath                  string   `yaml:"dbPath"`
	SpotifyId               string   `yaml:"-"`
	SpotifySecret           string   `yaml:"-"`
	path                    string   `yaml:"-"`
}

func (c *AppConfig) validate() error {
	if c.RunIntervalSeconds == 0 {
		return errors.New("appconfig: RunIntervalSeconds must be configured and more than 0")
	}

	if c.Port == 0 {
		return errors.New("appconfig: Port must be configured")
	}

	if c.WorkerCount == 0 {
		return errors.New("appconfig: WorkerCount must be configured and more than 0")
	}

	if c.WorkerTimeoutSeconds < 300 {
		return errors.New("appconfig: WorkerTimeoutSeconds must be configured and more than 300")
	}

	if c.SpotifyId == "" {
		return errors.New("appconfig: SpotifyId must be configured")
	}

	if c.SpotifySecret == "" {
		return errors.New("appconfig: SpotifySecret must be configured")
	}

	if c.SpotifyCallback == "" {
		return errors.New("appconfig: SpotifyCallback  must be configured")
	}

	return nil
}

type ConfigLoadError struct {
	Path string
	Err  error
}

func (e *ConfigLoadError) Error() string {
	return fmt.Sprintf("config: failed to load config at %s, error: %s", e.Path, e.Err.Error())
}

func Load(path string) (*AppConfig, error) {
	c := &AppConfig{IgnoreNotOwnedPlaylists: true, path: path, JsonDir: "json/", DbPath: "db/data.db"}

	err := loadYaml(c)
	if err != nil {
		return nil, &ConfigLoadError{Path: path, Err: err}
	}

	loadEnv(c)

	err = c.validate()
	if err != nil {
		return nil, &ConfigLoadError{Path: path, Err: err}
	}

	return c, nil
}

func loadYaml(c *AppConfig) error {
	if c.path == "" {
		return errors.New("no path provided")
	}

	data, err := ioutil.ReadFile(c.path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	return nil
}

func loadEnv(c *AppConfig) {
	c.SpotifyId = os.Getenv("SPOTIFY_ID")
	c.SpotifySecret = os.Getenv("SPOTIFY_SECRET")
}

// doesn't reload ENV based config values
func (c *AppConfig) Reload() (err error) {
	// this is good enough to make a copy because there are no deep refs
	newConf := (*c)
	err = loadYaml(&newConf)
	if err != nil {
		return
	}

	err = newConf.validate()
	if err != nil {
		return
	}

	// if new config valid and loaded update config values
	applyChangeableValues(&newConf, c)

	return
}

// updates current config with values found in newConf
// and saves it to disk
func (c *AppConfig) Update(newConf *AppConfig) (err error) {
	err = newConf.validate()
	if err != nil {
		return
	}

	oldCopy := (*c)

	applyChangeableValues(newConf, c)

	err = c.Persist()
	if err != nil {
		applyChangeableValues(&oldCopy, c)
		return
	}

	return
}

func applyChangeableValues(from *AppConfig, to *AppConfig) {
	to.RunIntervalSeconds = from.RunIntervalSeconds
	to.WorkerCount = from.WorkerCount
	to.WorkerTimeoutSeconds = from.WorkerTimeoutSeconds
	to.SavedPlaylistIds = from.SavedPlaylistIds
	to.IgnoredPlaylistIds = from.IgnoredPlaylistIds
	to.IgnoreNotOwnedPlaylists = from.IgnoreNotOwnedPlaylists
	to.IgnoreOwnedPlaylists = from.IgnoreOwnedPlaylists
}

// persists config on disk in multiple stages
// tries to create temp file, write to it then
// rename that to original location replacing
// previous version
func (c *AppConfig) Persist() (err error) {
	err = c.validate()
	if err != nil {
		return
	}

	// create temporary inside same directory as original
	// to avoid some issues with cross linking and/or perms
	f, err := os.CreateTemp(path.Dir(c.path), "tempconf")
	if err != nil {
		return
	}

	defer os.Remove(f.Name())

	cYaml, err := yaml.Marshal(c)

	err = os.WriteFile(f.Name(), cYaml, fs.ModeAppend)
	if err != nil {
		return
	}

	err = os.Rename(f.Name(), c.path)
	if err != nil {
		return
	}

	return
}
