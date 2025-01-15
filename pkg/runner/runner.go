package runner

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/tgagor/template-dockerfiles/pkg/cmd"
)

type Runner struct {
	tasks   []*cmd.Cmd
	threads int
	dryRun  bool
}

func New() *Runner {
	return &Runner{
		tasks:   []*cmd.Cmd{},
		dryRun:  false,
		threads: 1,
	}
}

func (r *Runner) Contains(task *cmd.Cmd) bool {
	for _, t := range r.tasks {
		if t.Equal(task) {
			return true
		}
	}
	return false
}

func (r *Runner) CountTasks() int {
	return len(r.tasks)
}

func (r *Runner) AddTask(task ...*cmd.Cmd) *Runner {
	r.tasks = append(r.tasks, task...)
	return r
}

func (r *Runner) AddUniq(task ...*cmd.Cmd) *Runner {
	// add only uniq calls
	for _, t := range task {
		if !r.Contains(t) {
			r.tasks = append(r.tasks, t)
		}
	}
	return r
}

func (r *Runner) DryRun(flag bool) *Runner {
	r.dryRun = flag
	return r
}

func (r *Runner) Threads(threads int) *Runner {
	r.threads = threads
	return r
}

func (r *Runner) GetTasks() []*cmd.Cmd {
	return r.tasks
}

// func (r Runner) Run() error {
// 	for _, c := range r.tasks {
// 		if r.dryRun {
// 			slog.Debug("DRY-RUN: Run", "cmd", c.String())
// 		} else {
// 			if err := c.Run(); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

func (r *Runner) Run() error {
	// Create a context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Workers get tasks from this channel
	tasks := make(chan *cmd.Cmd)

	// Feed the workers with tasks
	go func() {
		for _, c := range r.tasks {
			tasks <- c
		}
		// Workers will exit from range loop when channel is closed
		close(tasks)
	}()

	var wg sync.WaitGroup

	results := make(chan error)

	// use minimum required amount of workers
	threads := min(r.threads, len(r.tasks))
	log.Debug().Int("threads", threads).Int("max", max(r.threads, len(r.tasks))).Msg("Aquired parallelism")

	// Start the specified number of workers.
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range tasks {
				// Check if the context is canceled
				select {
				case <-ctx.Done():
					return // Stop processing tasks
				default:
				}

				// Execute the task
				if r.dryRun {
					log.Debug().Str("cmd", c.String()).Msg("DRY-RUN: Run")
				} else {
					if _, err := c.Run(ctx); err != nil {
						// Send the error to the results channel
						results <- err
						cancel() // Signal cancellation to all workers
						return   // Stop this worker
					}
				}
			}
		}()
	}

	// When workers are done, close results so that main will exit.
	go func() {
		wg.Wait()
		close(results)
	}()

	for err := range results {
		if err != nil {
			// cmd prints it anyway
			// log.Error().Err(err).Msg("Worker encountered error")
			return err
		}
	}

	return nil
}
