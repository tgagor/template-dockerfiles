package builder

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/cmd"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
	"github.com/tgagor/template-dockerfiles/pkg/runner"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

type DockerBuilder struct {
	flags  *config.Flags
	images []*image.Image

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

func (b *DockerBuilder) SetFlags(flags *config.Flags) {
	b.flags = flags
	b.SetThreads(flags.Threads)
	b.SetDryRun(!flags.Build)
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

func (b *DockerBuilder) Queue(image *image.Image) {
	b.images = append(b.images, image)
}

func (b *DockerBuilder) Build(img *image.Image) {
	builder := cmd.New("docker").Arg("build").
		Arg(optionsToArgs(img.Options)...).
		Arg("-f", img.Dockerfile).
		Arg("-t", img.UniqName()).
		Arg(labelsToArgs(img.Labels)...).
		Arg(buildArgsToArgs(img.BuildArgs)...).
		Arg(img.BuildContextDir).
		PreInfo("Building " + img.UniqName()).
		PostInfo("Built " + img.UniqName()).
		SetVerbose(b.flags.Verbose)
	b.buildTasks.AddTask(builder)
	b.Remove(img.UniqName()) // this image is temporary, remove it after build
}

// FIXME: images before squashing became unreferenced after squashing, so we should remove them
func (b *DockerBuilder) Squash(img *image.Image) {
	containerName := "run-" + util.SanitizeForFileName(img.UniqName())

	runItFirst := cmd.New("docker").Arg("run").
		Arg("--name", containerName).
		Arg(img.UniqName()).
		Arg("true").
		SetVerbose(b.flags.Verbose)
	b.squashRunImages.AddTask(runItFirst)

	imgMetadata, err := InspectImage(img.UniqName())
	util.FailOnError(err, "Couldn't inspect Docker image.")
	log.Trace().Interface("data", imgMetadata).Msg("Docker inspect result")
	b.imageSizesBefore[img.UniqName()] = imgMetadata[0].Size

	tmpTarFile := containerName + ".tar"
	exportIt := cmd.New("docker").Arg("export").
		Arg(containerName).
		Arg("-o", tmpTarFile).
		PreInfo(fmt.Sprintf("Squashing %s", img.UniqName())).
		SetVerbose(b.flags.Verbose)
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
			log.Error().Err(err).Str("image", img.UniqName()).Msg("Can't parse CMD")
		} else {
			importIt.Arg("--change", "CMD "+string(command))
		}

		// parsing VOLUME
		if vol, err := json.Marshal(item.Config.Volumes); err != nil {
			log.Error().Err(err).Str("image", img.UniqName()).Msg("Can't parse VOLUME")
		} else {
			importIt.Arg("--change", "VOLUME "+string(vol))
		}

		// parsing LABELS
		for key, value := range item.Config.Labels {
			importIt.Arg("--change", fmt.Sprintf("LABEL %s=\"%s\"", key, strings.ReplaceAll(value, "\n", "")))
		}

		// parsing ENTRYPOINT
		if entrypoint, err := json.Marshal(item.Config.Entrypoint); err != nil {
			log.Error().Err(err).Str("image", img.UniqName()).Msg("Can't parse ENTRYPOINT")
		} else {
			importIt.Arg("--change", "CMD "+string(entrypoint))
		}

		// parsing WORKDIR
		if item.Config.WorkingDir != "" {
			importIt.Arg("--change", "WORKDIR "+item.Config.WorkingDir)
		}
	}
	importIt.Arg(tmpTarFile).Arg(img.UniqName()).SetVerbose(b.flags.Verbose)
	b.squashImportTarsToImgs.AddTask(importIt)
	b.squashTempoaryTarFiles = append(b.squashTempoaryTarFiles, tmpTarFile)

	// remove interim images
	oldImgHash := strings.TrimPrefix(imgMetadata[0].Id, "sha256:")[:12]
	b.Remove(oldImgHash)
}

func (b *DockerBuilder) Tag(img *image.Image) {
	for _, tag := range img.Tags() {
		tagger := cmd.New("docker").Arg("tag").
			Arg(img.UniqName()).
			Arg(tag).
			PreInfo("Tagging " + tag).
			SetVerbose(b.flags.Verbose)
		b.taggingTasks.AddUniq(tagger)
	}
}

func (b *DockerBuilder) Push(img *image.Image) {
	for _, tag := range img.Tags() {
		pusher := cmd.New("docker").Arg("push").
			Arg(tag).
			PreInfo("Pushing " + tag)
		if !b.flags.Verbose {
			pusher.Arg("--quiet")
		}
		b.pushTasks.AddTask(pusher)
	}
}

func (b *DockerBuilder) Remove(imageName string) {
	remover := cmd.New("docker").Arg("image", "rm", "-f").
		Arg(imageName).
		SetVerbose(b.flags.Verbose)
	b.cleanupTasks.AddTask(remover)
}

// func (b *DockerBuilder) Run() error {
// 	// build images in parallel to fill the cache
// 	if b.flags.Build {
// 		for _, img := range b.images {
// 			b.Build(img)
// 		}

// 		if err := b.buildTasks.Run(); err != nil {
// 			return err
// 		}
// 	} else {
// 		log.Warn().Msg("Skipping building images. Use --build flag to build images.")
// 		return nil
// 	}

// 	// squash images
// 	if b.flags.Build && b.flags.Squash {
// 		for _, img := range b.images {
// 			b.Squash(img)
// 		}

// 		if err := b.RunSquashing(); err != nil {
// 			return err
// 		}
// 	}

// 	// tag single threaded
// 	if b.flags.Build {
// 		for _, img := range b.images {
// 			b.Tag(img)
// 		}

// 		if err := b.taggingTasks.Run(); err != nil {
// 			return err
// 		}
// 	}

// 	// push multi-threaded
// 	if b.flags.Push {
// 		for _, img := range b.images {
// 			b.Push(img)
// 		}

// 		if err := b.pushTasks.Run(); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (b *DockerBuilder) RunBuilding() error {
	if b.flags.Build {
		for _, img := range b.images {
			b.Build(img)
		}

		if err := b.buildTasks.Run(); err != nil {
			return err
		}
	} else {
		log.Warn().Msg("Skipping building images. Use --build flag to build images.")
		return nil
	}
	return nil
}

func (b *DockerBuilder) RunSquashing() error {
	// squash images
	if b.flags.Build && b.flags.Squash {
		for _, img := range b.images {
			b.Squash(img)
		}

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
	}

	return nil
}

func (b *DockerBuilder) RunTagging() error {
	// single threaded tagging
	if b.flags.Build {
		for _, img := range b.images {
			b.Tag(img)
		}

		if err := b.taggingTasks.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (b *DockerBuilder) RunPushing() error {
	// multi-threaded push
	if b.flags.Push {
		for _, img := range b.images {
			b.Push(img)
		}

		if err := b.pushTasks.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (b *DockerBuilder) Terminate() error {
	if b.flags.Build {
		if err := b.cleanupTasks.Run(); err != nil {
			return err
		}
	}

	// Cleanup temporary dockerfiles
	if b.flags.Delete {
		for _, img := range b.images {
			defer img.RemoveTemporaryDockerfile()
		}
	}

	return nil
}
