package image

import (
	"fmt"
	"maps"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

type Image struct {
	Name               string
	Registry           string
	Prefix             string
	DockerfileTemplate string
	Dockerfile         string
	BuildContextDir    string
	Variables          map[string]interface{}
	tags               []string
	Version            string
	Labels             map[string]string
	BuildArgs          map[string]string
	Platforms          []string
	Flags              *config.Flags
}

func New() *Image {
	return &Image{
		tags:      []string{},
		Labels:    map[string]string{},
		BuildArgs: map[string]string{},
		Platforms: []string{},
		Variables: map[string]interface{}{},
	}
}

func From(name string, cfg *config.Config, configSet map[string]interface{}, flags *config.Flags) *Image {
	img := New()

	img.Name = name
	img.Registry = cfg.Registry
	img.Prefix = cfg.Prefix
	img.Flags = flags
	img.Version = flags.Tag
	maps.Copy(img.Variables, configSet)
	img.updatePlatforms(cfg.GlobalPlatforms).
		updatePlatforms(cfg.Images[name].Platforms)

	// collect tags
	img.tags = append(img.tags, cfg.Images[name].Tags...) // non templated yet

	// collect labels
	maps.Copy(img.Labels, cfg.GlobalLabels)
	maps.Copy(img.Labels, collectOCILabels(img.ConfigSet()))
	img.SetMaintainer(cfg.Maintainer)
	if flags.Tag != "" {
		img.Labels["org.opencontainers.image.version"] = flags.Tag
	}
	maps.Copy(img.Labels, cfg.Images[name].Labels) // non templated yet

	// collect build arguments
	maps.Copy(img.BuildArgs, cfg.Images[name].BuildArgs)

	return img
}

func (i *Image) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("image name is required")
	}
	if i.Dockerfile == "" {
		return fmt.Errorf("Dockerfile is required")
	}
	if i.BuildContextDir == "" {
		return fmt.Errorf("BuildContextDir is required")
	}
	if len(i.Platforms) > 0 && i.Flags.Engine != "buildx" && i.Flags.Build {
		return fmt.Errorf("engine '%s' do not support multi-platform builds, use 'buildx' instead", i.Flags.Engine)
	}

	// check if users don't try to override reserved keys
	for k := range i.Variables {
		if isReservedKey(k) {
			return fmt.Errorf("variable key '%s' is reserved and cannot be used as variable", k)
		}
	}

	// validate if only allowed platforms are used
	for _, p := range i.Platforms {
		if !isAllowedPlatform(p) {
			return fmt.Errorf("platform '%s' is not allowed", p)
		}
	}

	// template templatedTags
	if templatedTags, err := TemplateList(i.tags, i.ConfigSet()); err != nil {
		return err
	} else {
		i.tags = templatedTags
	}

	// validate tags
	if len(i.tags) < 1 {
		// log.Error().Str("image", imageName).Msg("No 'tags' defined for")
		return fmt.Errorf("no 'tags' defined for %s - add 'tags' block to continue", i.Name)
	}

	// template labels
	if templatedLabels, err := TemplateMap(i.Labels, i.ConfigSet()); err != nil {
		return err
	} else {
		maps.Copy(i.Labels, templatedLabels)
	}

	// template build args
	if templatedBuildArgs, err := TemplateMap(i.BuildArgs, i.ConfigSet()); err != nil {
		return err
	} else {
		maps.Copy(i.BuildArgs, templatedBuildArgs)
	}

	// template Dockerfile
	log.Debug().Str("dockerfile", i.Dockerfile).Msg("Generating temporary")
	if err := TemplateFile(i.DockerfileTemplate, i.Dockerfile, i.ConfigSet()); err != nil {
		log.Error().Err(err).Str("dockerfile", i.Dockerfile).Msg("Failed to template Dockerfile")
		return err
	}

	return nil
}

func (i *Image) ConfigSet() map[string]interface{} {
	configSet := make(map[string]interface{})
	configSet["image"] = i.Name
	configSet["registry"] = i.Registry
	configSet["prefix"] = i.Prefix
	configSet["maintainer"] = i.Labels["maintainer"]
	configSet["platforms"] = i.Platforms
	maps.Copy(configSet, i.Variables)
	configSet["env"] = EnvVariables()
	configSet["tag"] = i.Version
	configSet["tags"] = i.tags
	configSet["labels"] = i.Labels
	configSet["args"] = i.BuildArgs

	log.Trace().Interface("config set", configSet).Msg("Generated")
	return configSet
}

func (i *Image) Representation() map[string]interface{} {
	return map[string]interface{}{
		"name":          i.Name,
		"registry":      i.Registry,
		"prefix":        i.Prefix,
		"dockerfile":    i.Dockerfile,
		"build_context": i.BuildContextDir,
		"variables":     i.Variables,
		// "tags":            i.tags,
		"version": i.Version,
		// "labels":          i.Labels,
		"build_args": i.BuildArgs,
		"platforms":  i.Platforms,
	}
}

func (i *Image) String() string {
	return i.Name
}

func (i *Image) UniqName() string {
	// name is required to avoid collisions between images or
	// when variables are not defined to have actual image name
	// ERROR: invalid tag "timezone-UTC": repository name must be lowercase
	return strings.ToLower(strings.Trim(fmt.Sprintf("%s-%s", i.Name, generateCombinationString(i.ConfigSet())), "-"))
}

func (i *Image) SetFlags(flags *config.Flags) *Image {
	i.Flags = flags
	return i
}

func (i *Image) SetName(name string) *Image {
	i.Name = name
	return i
}

// func (i *Image) FullName() string {
// 	return strings.ToLower(path.Join(i.Registry, i.Prefix, i.Name))
// }

func (i *Image) SetMaintainer(maintainer string) *Image {
	if maintainer != "" {
		i.Labels["maintainer"] = maintainer
		i.Labels["org.opencontainers.image.authors"] = maintainer
	} else {
		delete(i.Labels, "maintainer")
		delete(i.Labels, "org.opencontainers.image.authors")
	}
	return i
}

func (i *Image) updatePlatforms(platforms []string) *Image {
	if len(platforms) > 0 {
		i.Platforms = platforms
	}
	return i
}

func (i *Image) SetDockerfileTemplate(templateFile string) *Image {
	log.Debug().Str("dockerfile", templateFile).Msg("Processing")
	i.DockerfileTemplate = templateFile

	if strings.HasSuffix(templateFile, ".tpl") {
		i.Dockerfile = i.generateDockerfilePath()
	} else {
		i.Dockerfile = i.DockerfileTemplate
	}
	i.BuildContextDir = filepath.Dir(i.DockerfileTemplate)
	return i
}

func (i *Image) RemoveTemporaryDockerfile() {
	if i.Dockerfile != "" && i.Dockerfile != i.DockerfileTemplate {
		util.RemoveFile(i.Dockerfile)
	}
}

func (i *Image) Tags() []string {
	tags := []string{}
	for _, tag := range i.tags {
		tags = append(tags, strings.ToLower(path.Join(i.Registry, i.Prefix, tag)))
	}
	return tags
}

func (i *Image) Equal(image *Image) bool {
	return reflect.DeepEqual(i, image)
}

func isReservedKey(key string) bool {
	switch key {
	case
		"image",
		"registry",
		"prefix",
		"maintainer",
		"tag",
		"tags",
		"labels",
		"args",
		"env",
		"platforms":
		return true
	}
	return false
}

func isAllowedPlatform(platform string) bool {
	switch platform {
	case
		// following https://github.com/tonistiigi/binfmt
		"linux/amd64",
		"linux/arm64",
		"linux/riscv64",
		"linux/ppc64le",
		"linux/s390x",
		"linux/386",
		"linux/arm/v7",
		"linux/arm/v6":
		return true
	}
	return false
}

func generateCombinationString(configSet map[string]interface{}) string {
	var parts []string
	for k, v := range configSet {
		if !isReservedKey(k) {
			// Apply sanitization to both key and value
			safeKey := sanitizeForTag(k)
			safeValue := sanitizeForTag(fmt.Sprintf("%v", v))
			parts = append(parts, fmt.Sprintf("%s-%s", safeKey, safeValue))
			log.Debug().Str("key", safeKey).Str("value", safeValue).Msg("Combining")
		}
	}
	sort.Strings(parts)
	return strings.Trim(strings.Join(parts, "-"), "-")
}

// to avoid tags like this
// ERROR: invalid tag "test-case-7-alpine-3.21-crazy-map_key2_value2_-timezone-utc": invalid reference format
func sanitizeForTag(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	// FIXME: This can actually result in collisions if someone uses a lot of symbols in variables
	// 		  But I didn't face it yet, maybe it's not a problem at all
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
	return strings.Trim(reg.ReplaceAllString(input, "-"), "-")
}

func sanitizeForFileName(input string) string {
	// Replace any character that is not a letter, number, or safe symbol (-, _) with an underscore
	// FIXME: This can actually result in collisions if someone uses a lot of symbols in variables
	// 		  But I didn't face it yet, maybe it's not a problem at all
	reg := regexp.MustCompile(`[^a-zA-Z0-9-_\.]+`)
	return reg.ReplaceAllString(input, "_")
}

func (i *Image) generateDockerfilePath() string {
	dirname := filepath.Dir(i.DockerfileTemplate)
	filename := strings.Trim(fmt.Sprintf("%s-%s.Dockerfile", i.Name, generateCombinationString(i.ConfigSet())), "-")
	return filepath.Join(dirname, sanitizeForFileName(filename))
}
