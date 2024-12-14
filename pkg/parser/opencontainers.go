package parser

import (
	"fmt"
	"log/slog"
	"time"

	git "github.com/go-git/go-git/v5"
)

func getOCILabels(cfg map[string]interface{}) []string {
	labels := []string{}

	if cfg["maintainer"] != "" {
		labels = append(labels, "--label", fmt.Sprintf("maintainer=%s", cfg["maintainer"]))
	}

	if cfg["tag"] != "" {
		labels = append(labels, "--label", fmt.Sprintf("org.opencontainers.image.version=%s", cfg["tag"]))
	}

	timestamp := time.Now().Format(time.RFC3339)
	labels = append(labels, "--label", fmt.Sprintf("org.opencontainers.image.created=%s", timestamp))

	originUrl, hexsha, branch, err := readGitRepo(".")
	if err != nil {
		slog.Warn("Not being able to read git repo metadata, or not a git repo. Skipping.", "error", err)
	} else {
		if originUrl != "" {
			labels = append(labels, "--label", fmt.Sprintf("org.opencontainers.image.source=%s", originUrl))
		}
		if hexsha != "" {
			labels = append(labels, "--label", fmt.Sprintf("org.opencontainers.image.revision=%s", hexsha))
		}
		if branch != "" {
			labels = append(labels, "--label", fmt.Sprintf("org.opencontainers.image.branch=%s", branch))
		}
	}

	slog.Debug("Adding OCI", "labels", labels)
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
