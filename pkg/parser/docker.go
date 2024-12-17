package parser

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

type DockerInspect []struct {
	Config struct {
		Env        []string            `json:"Env"`
		Cmd        []string            `json:"Cmd"`
		Volumes    map[string]struct{} `json:"Volumes"`
		WorkingDir string              `json:"WorkingDir"`
		Entrypoint []string            `json:"Entrypoint"`
		Labels     map[string]string   `json:"Labels"`
	} `json:"Config"`
	Size uint64
}

func inspectImg(image string) (DockerInspect, error) {
	out, err := cmd.New("docker").Arg("inspect").Arg("--format").Arg("json").Arg(image).Output()
	util.FailOnError(err)

	// Create a variable to hold the unmarshaled data
	var inspect DockerInspect

	// Unmarshal the JSON data into the DockerInspect structure

	if err := json.Unmarshal([]byte(out), &inspect); err != nil {
		log.Error().Err(err).Msg("Error parsing JSON.")
		os.Exit(1)
	}

	return inspect, nil
}

// func imgSize(image string) string {
// 	out, err := cmd.New("docker").Arg("inspect").Arg("--format").Arg("json").Arg(image).Output()
// 	util.FailOnError(err)
// }
