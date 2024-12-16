package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Registry     string                 `yaml:"registry"`
	Prefix       string                 `yaml:"prefix"`
	Maintainer   string                 `yaml:"maintainer"`
	GlobalLabels map[string]string      `yaml:"labels"`
	Images       map[string]ImageConfig `yaml:"images"`
}

type ImageConfig struct {
	Dockerfile string                   `yaml:"dockerfile"`
	Variables  map[string][]interface{} `yaml:"variables"`
	Excludes   []map[string]string      `yaml:"excludes"`
	Tags       []string                 `yaml:"tags"`
	Labels     map[string]string        `yaml:"labels"`
}

func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		slog.Error("Error loading config", "error", err)
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		slog.Error("Decoding YAML "+filename+" failed! Check syntax and try again", "error", err)
		return nil, err
	}
	return &cfg, nil
}
