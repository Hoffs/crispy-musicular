package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	RunIntervalSeconds uint64 `yaml:"runIntervalSeconds"`
	Port               uint32 `yaml:"port"`
	SpotifyCallback    string `yaml:"spotifyCallback"`
	SpotifyId          string
	SpotifySecret      string
}

func (c *AppConfig) validate() error {
	if c.RunIntervalSeconds < 300 {
		return errors.New(fmt.Sprintf("appconfig: RunIntervalSeconds must be more than 300, received: %d", c.RunIntervalSeconds))
	}

	if c.Port == 0 {
		return errors.New(fmt.Sprint("appconfig: Port must be configured"))
	}

	if c.SpotifyId == "" {
		return errors.New(fmt.Sprint("appconfig: SpotifyId must be configured"))
	}

	if c.SpotifySecret == "" {
		return errors.New(fmt.Sprint("appconfig: SpotifySecret must be configured"))
	}

	if c.SpotifyCallback == "" {
		return errors.New(fmt.Sprint("appconfig: SpotifyCallback  must be configured"))
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
	c := &AppConfig{}

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
