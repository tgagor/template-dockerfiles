package config

import (
	"io"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Registry        string                 `yaml:"registry"`
	Prefix          string                 `yaml:"prefix"`
	Maintainer      string                 `yaml:"maintainer"`
	GlobalLabels    map[string]string      `yaml:"labels"`
	GlobalPlatforms []string               `yaml:"platforms"`
	GlobalOptions   []string               `yaml:"options"`
	Images          map[string]ImageConfig `yaml:"images"`
	ImageOrder      []string               `yaml:"-"` // To preserve the order of images
}

type imageLoader struct {
	Images yaml.Node `yaml:"images"`
}

type ImageConfig struct {
	Dockerfile string                   `yaml:"dockerfile"`
	Variables  map[string]interface{}   `yaml:"variables"` // Changed to interface{}
	Excludes   []map[string]interface{} `yaml:"excludes"`
	Tags       []string                 `yaml:"tags"`
	Labels     map[string]string        `yaml:"labels"`
	BuildArgs  map[string]string        `yaml:"args"`
	Platforms  []string                 `yaml:"platforms"`
	Options    []string                 `yaml:"options"`
}

func Load(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Err(err).Msg("Error loading config")
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		log.Error().Err(err).Msg("Decoding YAML " + filename + " failed! Check syntax and try again")
		return nil, err
	}

	// Seek to the beginning of the file to read the image order
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Error().Err(err).Msg("Error seeking file")
		return nil, err
	}
	// Preserve the order of images
	var loader imageLoader
	if err := yaml.NewDecoder(file).Decode(&loader); err != nil {
		log.Error().Err(err).Msg("Decoding YAML " + filename + " failed! Check syntax and try again")
		return nil, err
	}
	cfg.ImageOrder = []string{}
	for _, node := range loader.Images.Content {
		if node.Tag == "!!str" {
			cfg.ImageOrder = append(cfg.ImageOrder, node.Value)
		}
	}
	log.Debug().Interface("Config", cfg).Msg("Config loaded")

	return &cfg, nil
}
