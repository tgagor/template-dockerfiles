package config

import (
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Registry     string                 `yaml:"registry"`
	Prefix       string                 `yaml:"prefix"`
	Maintainer   string                 `yaml:"maintainer"`
	GlobalLabels map[string]string      `yaml:"labels"`
	Images       map[string]ImageConfig `yaml:"images"`
	ImageOrder   []string               `yaml:"-"` // To preserve the order of images
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
		log.Error().Err(err).Msg("Error loading config")
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		log.Error().Err(err).Msg("Decoding YAML " + filename + " failed! Check syntax and try again")
		return nil, err
	}

	// Preserve the order of images
	var rawYaml map[string]interface{}
	if err := yaml.NewDecoder(file).Decode(&rawYaml); err != nil {
		log.Error().Err(err).Msg("Decoding raw YAML " + filename + " failed! Check syntax and try again")
		return nil, err
	}
	if images, ok := rawYaml["images"].(map[string]interface{}); ok {
		for key := range images {
			cfg.ImageOrder = append(cfg.ImageOrder, key)
		}
	}

	return &cfg, nil
}
