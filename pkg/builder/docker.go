package builder

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

type DockerBuilder struct {
	buildTasks   *runner.Runner
	taggingTasks *runner.Runner
	pushTasks    *runner.Runner
	cleanupTasks *runner.Runner

	// just for squashing
	squashRunImages        *runner.Runner
	squashExportImages     *runner.Runner
	squashImportTarsToImgs *runner.Runner
	squashTempoaryTarFiles []string
	imageSizesBefore       map[string]uint64
}

func (b *DockerBuilder) Init() error {
	b.buildTasks = runner.New()
	b.taggingTasks = runner.New()
	b.pushTasks = runner.New()
	b.cleanupTasks = runner.New()

	b.squashRunImages = runner.New()
	b.squashExportImages = runner.New()
	b.squashImportTarsToImgs = runner.New()
	b.squashTempoaryTarFiles = []string{}
	b.imageSizesBefore = map[string]uint64{}

	log.Info().Str("engine", "docker").Msg("Initializing")
	return nil
}

func (b *DockerBuilder) SetThreads(threads int) {
	b.buildTasks.Threads(threads)
	// b.tagTasks have to use 1 thread
	b.pushTasks.Threads(threads)
	b.cleanupTasks.Threads(threads)

	b.squashRunImages.Threads(threads)
	b.squashExportImages.Threads(threads)
	b.squashImportTarsToImgs.Threads(threads)
}

func (b *DockerBuilder) SetDryRun(dryRun bool) {
	b.buildTasks.DryRun(dryRun)
	b.taggingTasks.DryRun(dryRun)
	b.pushTasks.DryRun(dryRun)
	b.cleanupTasks.DryRun(dryRun)

	b.squashRunImages.DryRun(dryRun)
	b.squashExportImages.DryRun(dryRun)
	b.squashImportTarsToImgs.DryRun(dryRun)
}

func (b *DockerBuilder) Build(dockerfile, imageName string, configSet map[string]interface{}, contextDir string, verbose bool) {
	builder := cmd.New("docker").Arg("build").
		Arg("-f", dockerfile).
		Arg("-t", imageName).
		Arg(labelsToArgs(configSet["labels"].(map[string]string))...).
		Arg(buildArgsToArgs(configSet["args"].(map[string]string))...).
		Arg(contextDir).
		PreInfo("Building " + imageName).
		SetVerbose(verbose)
	b.buildTasks.AddTask(builder)
}

// TODO: should I add a flag for original image removal?
//
//	they're unreferenced after squashing
func (b *DockerBuilder) Squash(imageName string, verbose bool) {
	containerName := "run-" + sanitizeForFileName(imageName)

	runItFirst := cmd.New("docker").Arg("run").
		Arg("--name", containerName).
		Arg(imageName).
		Arg("true").
		SetVerbose(verbose)
	b.squashRunImages.AddTask(runItFirst)

	imgMetadata, err := InspectImage(imageName)
	util.FailOnError(err, "Couldn't inspect Docker image.")
	log.Debug().Interface("data", imgMetadata).Msg("Docker inspect result")
	b.imageSizesBefore[imageName] = imgMetadata[0].Size

	tmpTarFile := containerName + ".tar"
	exportIt := cmd.New("docker").Arg("export").
		Arg(containerName).
		Arg("-o", tmpTarFile).
		PreInfo(fmt.Sprintf("Squashing %s", imageName)).
		SetVerbose(verbose)
	b.squashExportImages.AddTask(exportIt)
	b.cleanupTasks.AddTask(cmd.New("docker").Arg("rm").Arg("-f").Arg(containerName))

	importIt := cmd.New("docker").Arg("import")
	for _, item := range imgMetadata {
		// paring ENV
		for _, env := range item.Config.Env {
			importIt.Arg("--change", "ENV "+env)
		}

		// parsing CMD
		if command, err := json.Marshal(item.Config.Cmd); err != nil {
			log.Error().Err(err).Str("image", imageName).Msg("Can't parse CMD")
		} else {
			importIt.Arg("--change", "CMD "+string(command))
		}

		// parsing VOLUME
		if vol, err := json.Marshal(item.Config.Volumes); err != nil {
			log.Error().Err(err).Str("image", imageName).Msg("Can't parse VOLUME")
		} else {
			importIt.Arg("--change", "VOLUME "+string(vol))
		}

		// parsing LABELS
		for key, value := range item.Config.Labels {
			importIt.Arg("--change", fmt.Sprintf("LABEL %s=\"%s\"", key, strings.ReplaceAll(value, "\n", "")))
		}

		// parsing ENTRYPOINT
		if entrypoint, err := json.Marshal(item.Config.Entrypoint); err != nil {
			log.Error().Err(err).Str("image", imageName).Msg("Can't parse ENTRYPOINT")
		} else {
			importIt.Arg("--change", "CMD "+string(entrypoint))
		}

		// parsing WORKDIR
		if item.Config.WorkingDir != "" {
			importIt.Arg("--change", "WORKDIR "+item.Config.WorkingDir)
		}
	}
	importIt.Arg(tmpTarFile).Arg(imageName).SetVerbose(verbose)
	b.squashImportTarsToImgs.AddTask(importIt)
	b.squashTempoaryTarFiles = append(b.squashTempoaryTarFiles, tmpTarFile)

	// remove interim images
	oldImgHash := strings.TrimPrefix(imgMetadata[0].Id, "sha256:")[:12]
	b.Remove(oldImgHash, verbose)
}

func (b *DockerBuilder) Tag(imageName, taggedImage string, verbose bool) {
	tagger := cmd.New("docker").Arg("tag").
		Arg(imageName).
		Arg(taggedImage).
		PreInfo("Tagging " + taggedImage).
		SetVerbose(verbose)
	b.taggingTasks.AddUniq(tagger)
}

func (b *DockerBuilder) Push(taggedImage string, verbose bool) {
	pusher := cmd.New("docker").Arg("push").
		Arg(taggedImage).
		PreInfo("Pushing " + taggedImage)
	if !verbose {
		pusher.Arg("--quiet")
	}
	b.pushTasks.AddTask(pusher)
}

func (b *DockerBuilder) Remove(imageName string, verbose bool) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(verbose)
	b.cleanupTasks.AddTask(remover)
}

func (b *DockerBuilder) RunBuilding() error {
	return b.buildTasks.Run()
}

func (b *DockerBuilder) RunSquashing() error {
	defer util.RemoveFile(b.squashTempoaryTarFiles...)

	if err := b.squashRunImages.Run(); err != nil {
		return err
	}
	if err := b.squashExportImages.Run(); err != nil {
		return err
	}
	if err := b.squashImportTarsToImgs.Run(); err != nil {
		return err
	}

	for imageName, sizeBefore := range b.imageSizesBefore {
		log.Debug().Str("image", imageName).Uint64("sizeBefore", sizeBefore).Msg("Squashing")
		imgMetadata, err := InspectImage(imageName)
		if err != nil {
			return err
		}
		sizeAfter := imgMetadata[0].Size
		percentage := float64(sizeAfter)*100/float64(sizeBefore) - 100
		log.Info().Str("image", imageName).Str("was", util.ByteCountIEC(sizeBefore)).Str("is", util.ByteCountIEC(sizeAfter)).Str("reduction", fmt.Sprintf("%.1f%%", percentage)).Msg("Squashed")
	}

	return nil
}
func (b *DockerBuilder) RunTagging() error {
	return b.taggingTasks.Run()
}
func (b *DockerBuilder) RunPushing() error {
	return b.pushTasks.Run()
}
func (b *DockerBuilder) RunCleanup() error {
	return b.cleanupTasks.Run()
}

func (b *DockerBuilder) Shutdown() error {
	return nil
}
