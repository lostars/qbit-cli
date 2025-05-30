package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var (
	CfgPath string
	config  *Config
	Debug   bool
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
		DefaultSearchPlugin string `yaml:"default-search-plugin"`
	} `yaml:"torrent"`

	Jackett struct {
		Host   string `yaml:"host" validate:"required"`
		ApiKey string `yaml:"api-key"`
		Cookie string `yaml:"cookie"`
	} `yaml:"jackett"`

	Emby struct {
		Host   string `yaml:"host"`
		ApiKey string `yaml:"api-key"`
		User   string `yaml:"user"`
	}

	NeteaseMusicCookie string `yaml:"netease_music_cookie"`
	QQMusicCookie      string `yaml:"qq_music_cookie"`
}

func loadDefaultConfig() []byte {
	home, _ := os.UserHomeDir()
	if home != "" {
		// load from user home .config/qbit-cli/config.yaml
		CfgPath = filepath.Join(home, ".config", "qbit-cli", "config.yaml")
		file, _ := os.ReadFile(CfgPath)
		if file != nil {
			return file
		}
	}
	execPath, _ := os.Executable()
	if execPath != "" {
		CfgPath = filepath.Join(filepath.Dir(execPath), "config.yaml")
		file, _ := os.ReadFile(CfgPath)
		if file != nil {
			return file
		}
	}
	return nil
}

func GetConfig() *Config {
	if config != nil {
		return config
	}
	var file []byte
	if CfgPath == "" {
		file = loadDefaultConfig()
		if file == nil {
			panic("default config file load failed")
		}
	} else {
		f, err := os.ReadFile(CfgPath)
		if err != nil {
			panic(err.Error())
		}
		file = f
	}

	cfg := Config{}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		panic(err.Error())
	}
	return &cfg
}
