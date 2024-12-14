package parser

import (
	"fmt"
	"log/slog"
)

func getOCILabels(cfg map[string]interface{}) []string {
	labels := []string{}

	if cfg["maintainer"] != "" {
		labels = append(labels, "--label", fmt.Sprintf("maintainer=\"%s\"", cfg["maintainer"]))
	}

	if cfg["tag"] != "" {
		labels = append(labels, "--label", fmt.Sprintf("org.opencontainers.image.version=%s", cfg["tag"]))
	}

	slog.Debug("Adding OCI", "labels", labels)
	return labels
}

// def get_opencontainer_labels(playbook):
//     labels = []

//     maintainer = playbook.get("maintainer")
//     if maintainer and maintainer.strip():
//         labels.extend(["--label", f"maintainer={maintainer}"])

//     if flags.TAG:
//         labels.extend(["--label", f"org.opencontainers.image.version={flags.TAG}"])

//     repo = git.Repo(search_parent_directories=True)
//     try:
//         labels.extend(
//             ["--label", f"org.opencontainers.image.source={repo.remotes.origin.url}"]
//         )
//     except AttributeError:
//         pass

//     try:
//         labels.extend(
//             ["--label", f"org.opencontainers.image.revision={repo.head.object.hexsha}"]
//         )
//     except AttributeError:
//         pass

//     try:
//         labels.extend(
//             ["--label", f"org.opencontainers.image.branch={repo.active_branch}"]
//         )
//     except AttributeError:
//         pass

//     n = datetime.datetime.now(datetime.timezone.utc)
//     labels.extend(["--label", f"org.opencontainers.image.created={n.isoformat()}"])

//     return labels
