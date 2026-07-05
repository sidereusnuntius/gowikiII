package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/sidereusnuntius/gowiki/internal/wikilog"
)

type WikiConfig struct {
	Name        string   `toml:"name"`
	Host        string   `toml:"host"`
	Port        int      `toml:"port"`
	URL         *url.URL `toml:"-"`
	SharedInbox string   `toml:"-"`
	Testing     bool     `toml:"testing"`
	Database    DbConfig `toml:"database"`
}

type DbConfig struct {
	URL     string `toml:"url"`
	Timeout int    `toml:"timeout"`
	Test    bool   `toml:"-"`
	// TODO: add other options
}

func DefaultConfig(addr string) WikiConfig {
	url, _ := url.Parse(addr)
	port, _ := strconv.Atoi(url.Port())
	return WikiConfig{
		Name:        url.Host,
		Host:        url.Host,
		Port:        port,
		URL:         url,
		SharedInbox: url.JoinPath("inbox").String(),
		Database: DbConfig{
			URL: "./test.db",
		},
	}
}

func ReadConfig() (config WikiConfig, err error) {
	config = DefaultConfig("http://localhost:8080")

	// Try to find file to read config from, if it exists.
	if configpath := os.Getenv("CONFIG_FILEPATH"); configpath != "" {
		wikilog.Logger.Debug().
			Str("path", configpath).
			Msg("CONFIG_FILEPATH set; trying to read config file")
		config, err = readConfigFile(configpath)
		if err != nil {
			return WikiConfig{}, err
		}
	} else {
		wikilog.Logger.Debug().
			Msg("using default config")
	}

	config.Name = readVarWithDefault(config.Name, "WIKI_NAME")
	config.Host = readVarWithDefault(config.Name, "WIKI_HOST")
	config.Port = intVarWithDefault(config.Port, "WIKI_PORT")
	config.Database.URL = readVarWithDefault(config.Database.URL, "WIKI_DB_URL")

	wikiurl := fmt.Sprintf("http://%s:%d", config.Host, config.Port)
	config.URL, err = url.Parse(wikiurl)
	if err != nil {
		wikilog.Logger.Error().
			Err(err).
			Str("url", wikiurl).
			Msg("invalid URL for wiki")
	}

	config.SharedInbox = config.URL.JoinPath("inbox").String()

	return config, nil
}

func readConfigFile(path string) (WikiConfig, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			wikilog.Logger.Debug().
				Str("path", path).
				Msg("config file not found, using default config")
			return WikiConfig{}, err
		}

		wikilog.Logger.Error().
			Err(err).
			Str("path", path).
			Msg("failed to look up file")
		return WikiConfig{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		wikilog.Logger.Error().
			Err(err).
			Str("path", path).
			Msg("failed to read config file")
	}

	var config WikiConfig
	_, err = toml.Decode(string(content), &config)
	if err != nil {
		wikilog.Logger.Error().
			Err(err).
			Str("path", path).
			Msg("failed to decode the config file")
		return WikiConfig{}, err
	}

	return config, nil
}

func readVarWithDefault(defaultVal, varname string) string {
	v, ok := os.LookupEnv(varname)
	if ok {
		return v
	}

	return defaultVal
}

func intVarWithDefault(defaultValue int, varname string) int {
	val, ok := os.LookupEnv(varname)
	if !ok {
		return defaultValue
	}

	valint, err := strconv.Atoi(val)
	if err != nil {
		wikilog.Logger.Error().
			Err(err).
			Str("var", val).
			Msg("cannot parse variable as integer")
		return defaultValue
	}

	return valint
}
