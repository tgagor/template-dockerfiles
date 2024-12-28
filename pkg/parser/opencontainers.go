package parser

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	git "github.com/go-git/go-git/v5"
)

// Follow:
// https://github.com/opencontainers/image-spec/blob/main/annotations.md
func collectOCILabels(cfg map[string]interface{}) map[string]string {
	labels := map[string]string{}

	if cfg["maintainer"] != "" {
		labels["maintainer"] = cfg["maintainer"].(string)
		labels["org.opencontainers.image.authors"] = cfg["maintainer"].(string)
	}

	if cfg["tag"] != "" {
		labels["org.opencontainers.image.version"] = cfg["tag"].(string)
	}

	timestamp := time.Now().Format(time.RFC3339)
	labels["org.opencontainers.image.created"] = string(timestamp)

	originUrl, hexsha, branch, err := readGitRepo(".")
	if err != nil {
		log.Warn().Err(err).Msg("Not being able to read git repo metadata, or not a git repo. Skipping.")
	} else {
		if originUrl != "" {
			labels["org.opencontainers.image.source"] = originUrl
		}
		if hexsha != "" {
			labels["org.opencontainers.image.revision"] = hexsha
		}
		if branch != "" {
			labels["org.opencontainers.image.branch"] = branch
		}
	}

	log.Debug().Interface("labels", labels).Msg("Adding OCI")
	return labels
}

func readGitRepo(path string) (originURL string, commitHex string, branchName string, err error) {
	// Open the local git repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			// Return nothing if it's not a Git repository
			return "", "", "", nil
		}
		return "", "", "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the repository's remotes and find the origin remote
	remotes, err := repo.Remotes()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to list remotes: %w", err)
	}

	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			if len(remote.Config().URLs) > 0 {
				originURL = remote.Config().URLs[0]
			}
			break
		}
	}

	// Get the HEAD reference (current branch or commit)
	head, err := repo.Head()
	if err != nil {
		return originURL, "", "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitHex = head.Hash().String()

	// Determine the current branch name
	if head.Name().IsBranch() {
		branchName = head.Name().Short()
	} else {
		branchName = "" // Detached HEAD state
	}

	return originURL, commitHex, branchName, nil
}
