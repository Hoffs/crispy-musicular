package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
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
	SpotifyId               string
	SpotifySecret           string
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
	c := &AppConfig{IgnoreNotOwnedPlaylists: true}

	err := loadYaml(c, path)
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

func loadYaml(c *AppConfig, path string) error {
	if path == "" {
		return errors.New("no path provided")
	}

	data, err := ioutil.ReadFile(path)
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
	_ = godotenv.Load()
	_ = godotenv.Load(".env.local")

	c.SpotifyId = os.Getenv("SPOTIFY_ID")
	c.SpotifySecret = os.Getenv("SPOTIFY_SECRET")
}
