package config

import (
	"fmt"
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

	Emby struct {
		Host   string `yaml:"host"`
		ApiKey string `yaml:"api-key"`
		User   string `yaml:"user"`
	}
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

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}
	var file []byte
	if CfgPath == "" {
		file = loadDefaultConfig()
		if file == nil {
			return nil, &CfgError{"default config load error", "", nil}
		}
	} else {
		f, err := os.ReadFile(CfgPath)
		if err != nil {
			return nil, &CfgError{"config load error", CfgPath, nil}
		}
		file = f
	}

	cfg := Config{}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, &CfgError{"config file parse error", CfgPath, err}
	}
	return &cfg, nil
}

type CfgError struct {
	message string
	path    string
	err     error
}

func (cfg *CfgError) Error() string {
	errStr := ""
	if cfg.err != nil {
		errStr = cfg.err.Error()
	}
	return fmt.Sprintf("%s: %s %s", cfg.message, cfg.path, errStr)
}
