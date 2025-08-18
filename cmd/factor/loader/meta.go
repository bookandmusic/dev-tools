package loader

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// PluginMeta represents the metadata of a plugin defined in meta.yml
type PluginMeta struct {
	Name        string             `yaml:"name"`
	Description string             `yaml:"description"`
	Type        string             `yaml:"type"`
	Version     string             `yaml:"version"`
	Commands    map[string]Command `yaml:"commands"`
}

// Command represents a command that a plugin can execute
type Command struct {
	Description string             `yaml:"description"`
	Usage       string             `yaml:"usage"`
	Options     []Option           `yaml:"options,omitempty"`
	Subcommands map[string]Command `yaml:"subcommands,omitempty"`
}

// Option represents a command-line option
type Option struct {
	Name        string `yaml:"name"`
	Short       string `yaml:"short,omitempty"`
	Description string `yaml:"description"`
	Value       string `yaml:"value,omitempty"`
}

func LoadPluginMeta(pluginPath string, info os.FileInfo) (*PluginMeta, error) {
	if !info.IsDir() {
		return nil, fmt.Errorf("plugin path is not a directory")
	}
	metaPath := filepath.Join(pluginPath, "meta.yml")
	if _, err := os.Stat(metaPath); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var meta PluginMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
