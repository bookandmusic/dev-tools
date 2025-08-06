package config

import (
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	RootDir string `yaml:"root_dir"`
	BinDir  string `yaml:"bin_dir"`
}

var Global *Config

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	path := os.Getenv("DEV_TOOLS_HOME")
	if path == "" {
		// 默认路径
		userHome, _ := os.UserHomeDir()
		path = filepath.Join(userHome, ".tools")
	}
	config := &Config{
		RootDir: path,
	}
	configPath := filepath.Join(path, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		Global = config
		return config, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		Global = config
		return config, err
	}

	Global = &cfg
	Global.RootDir = path
	return &cfg, nil
}
