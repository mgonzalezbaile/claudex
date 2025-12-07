// Package config provides configuration file loading and parsing for Claudex.
// It supports loading .claudex.toml files with options for documentation paths
// and file overwrite behavior.
package config

import (
	"github.com/BurntSushi/toml"
	"github.com/spf13/afero"
)

type Config struct {
	Doc         []string `toml:"doc"`
	NoOverwrite bool     `toml:"no_overwrite"`
}

// Load loads configuration from the specified path using the provided filesystem
func Load(fs afero.Fs, path string) (*Config, error) {
	config := &Config{
		Doc:         []string{},
		NoOverwrite: false,
	}

	if _, err := fs.Stat(path); err == nil {
		data, err := afero.ReadFile(fs, path)
		if err != nil {
			return nil, err
		}
		if _, err := toml.Decode(string(data), config); err != nil {
			return nil, err
		}
	}
	return config, nil
}
