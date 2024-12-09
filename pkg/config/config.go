package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Registry   string                 `yaml:"registry"`
	Prefix     string                 `yaml:"prefix"`
	Maintainer string                 `yaml:"maintainer"`
	Images     map[string]ImageConfig `yaml:"images"`
}

type ImageConfig struct {
	Dockerfile string                `yaml:"dockerfile"`
	Variables  map[string][]string   `yaml:"variables"`
	Excludes   []map[string]string   `yaml:"excludes"`
	Labels     []string              `yaml:"labels"`
}

func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
