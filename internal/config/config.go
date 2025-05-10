package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var (
	CfgPath string
	config  *Config
)

type Config struct {
	Server struct {
		Host     string `yaml:"host" validate:"required"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"server"`

	Torrent struct {
		DefaultSaveCategory string `yaml:"default-save-category"`
		DefaultSaveTags     string `yaml:"default-save-tags"`
		DefaultSavePath     string `yaml:"default-save-path"`
	} `yaml:"torrent"`

	Jackett struct {
		Host   string `yaml:"host" validate:"required"`
		ApiKey string `yaml:"api-key"`
	} `yaml:"jackett"`
}

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}
	if CfgPath == "" {
		// load from the path where the command exists
		execPath, _ := os.Executable()
		if execPath != "" {
			CfgPath = filepath.Join(filepath.Dir(execPath), "config.yaml")
		}
	}
	file, err := os.ReadFile(CfgPath)
	if err != nil {
		return nil, err
	}

	cfg := Config{}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *Config) ValidateJackettConfig() bool {
	return cfg.Jackett.ApiKey != "" && cfg.Jackett.Host != ""
}
