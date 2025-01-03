package image

import (
	"maps"
	"path"
	"reflect"
	"strings"

	"github.com/tgagor/template-dockerfiles/pkg/config"
)

type Image struct {
	Name            string
	Registry        string
	Prefix          string
	Dockerfile      string
	BuildContextDir string
	Variables       map[string]interface{}
	tags            []string
	Labels          map[string]string
	BuildArgs       map[string]string
	Platforms       []string
	Flags           *config.Flags
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

func From(configSet map[string]interface{}, flags *config.Flags) *Image {
	img := Image{
		// Name:      configSet["name"].(string),
		Registry: configSet["registry"].(string),
		Prefix:   configSet["prefix"].(string),
		// tags:      configSet["tags"].([]string),
		Labels:    configSet["labels"].(map[string]string),
		BuildArgs: configSet["args"].(map[string]string),
		// Platforms: configSet["platforms"].([]string),
		Flags: flags,
	}

	img.AddTags(configSet["tags"].([]string)...)
	img.SetMaintainer(configSet["maintainer"].(string))
	img.SetPlatforms(configSet["platforms"].([]string))

	return &img
}

func (i *Image) String() string {
	return i.Name
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

func (i *Image) SetPlatforms(platforms []string) *Image {
	if len(platforms) > 0 {
		i.Platforms = platforms
	}
	return i
}

func (i *Image) SetDockerfile(dockerfile string) *Image {
	i.Dockerfile = dockerfile
	return i
}

func (i *Image) SetBuildContextDir(contextdir string) *Image {
	i.BuildContextDir = contextdir
	return i
}

func (i *Image) AddTags(tags ...string) *Image {
	i.tags = append(i.tags, tags...)
	return i
}

func (i *Image) Tags() []string {
	tags := []string{}
	for _, tag := range i.tags {
		tags = append(tags, strings.ToLower(path.Join(i.Registry, i.Prefix, tag)))
	}
	return tags
}

func (i *Image) AddLabels(labels map[string]string) *Image {
	maps.Copy(i.Labels, labels)
	return i
}

func (i *Image) AddArgs(args map[string]string) *Image {
	maps.Copy(i.BuildArgs, args)
	return i
}

func (i *Image) AddVariables(variables map[string]interface{}) *Image {
	maps.Copy(i.Variables, variables)
	return i
}

func (i *Image) Equal(image *Image) bool {
	return reflect.DeepEqual(i, image)
}
