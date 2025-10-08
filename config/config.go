package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AbeEstrada/tuit/constants"
)

type ConfigAuth struct {
	Server       string `json:"server"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AccessToken  string `json:"access_token"`
}

type Config struct {
	Auth ConfigAuth `json:"auth"`
}

var configDirName = strings.ToLower(constants.AppName)
var configFileName = "config.json"

func GetConfigDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, configDirName)
		}
	}

	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, configDirName)
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".config", configDirName)
	}

	// Cannot determine config directory, using current directory
	return filepath.Join(".", configDirName)
}

func GetConfigFile() string {
	configDir := GetConfigDir()
	return filepath.Join(configDir, configFileName)
}

func LoadConfig() (*Config, error) {
	configFile := GetConfigFile()
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", configFile, err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON from %s: %w", configFile, err)
	}

	return &config, nil
}
