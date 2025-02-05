package builder

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
)

type DockerInspect []struct {
	Id     string `json:"Id"`
	Size   uint64 `json:"Size"`
	Config struct {
		Env        []string            `json:"Env"`
		Cmd        []string            `json:"Cmd"`
		Volumes    map[string]struct{} `json:"Volumes"`
		WorkingDir string              `json:"WorkingDir"`
		Entrypoint []string            `json:"Entrypoint"`
		Labels     map[string]string   `json:"Labels"`
	} `json:"Config"`
}

func InspectImage(image string) (DockerInspect, error) {
	out, err := cmd.New("docker").Arg("inspect").Arg("--format").Arg("json").Arg(image).Output()
	if err != nil {
		return nil, err
	}
	// Create a variable to hold the unmarshaled data
	var inspect DockerInspect

	log.Trace().Interface("output", out).Msg("Inspect output")

	// Unmarshal the JSON data into the DockerInspect structure
	if err := json.Unmarshal([]byte(out), &inspect); err != nil {
		log.Error().Err(err).Msg("Error parsing JSON.")
		return nil, err
	}

	return inspect, nil
}

func labelsToArgs(labels map[string]string) []string {
	args := []string{}
	for k, v := range labels {
		args = append(args, "--label", k+"="+v)
	}
	return args
}

func buildArgsToArgs(buildArgs map[string]string) []string {
	args := []string{}
	for k, v := range buildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%v", k, v))
	}
	return args
}

func platformsToArgs(platforms []string) []string {
	args := []string{}
	args = append(args, "--platform", strings.Join(platforms, ","))
	return args
}

func optionsToArgs(options map[string]string) []string {
	args := []string{}

	for opt, val := range options {
		if opt != "" && val != "" {
			args = append(args, "--"+opt, val)
		}
		if opt != "" && val == "" {
			args = append(args, "--"+opt)
		}
	}

	return args
}
