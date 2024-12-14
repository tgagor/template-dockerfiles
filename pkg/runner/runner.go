package runner

import (
	"log/slog"
	"sync"
	"td/pkg/cmd"
)

type Runner struct {
	tasks   []cmd.Cmd
	threads int
	dryRun  bool
}

func New() Runner {
	return Runner{
		tasks:   []cmd.Cmd{},
		dryRun:  false,
		threads: 1,
	}
}

func (r Runner) Contains(task cmd.Cmd) bool {
	for _, t := range r.tasks {
		if t.Equal(task) {
			return true
		}
	}
	return false
}

func (r Runner) AddTask(task ...cmd.Cmd) Runner {
	// add only uniq calls
	for _, t := range task {
		if !r.Contains(t) {
			r.tasks = append(r.tasks, t)
		}
	}
	return r
}

func (r Runner) DryRun(flag bool) Runner {
	r.dryRun = flag
	return r
}

func (r Runner) Threads(threads int) Runner {
	r.threads = threads
	return r
}

func (r Runner) Run() error {
	for _, c := range r.tasks {
		if r.dryRun {
			slog.Debug("DRY-RUN: Run", "cmd", c.String())
		} else {
			if err := c.Run(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r Runner) RunParallel() error {
	// Workers get tasks from this channel
	tasks := make(chan cmd.Cmd)

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
	slog.Debug("Aquired parallelism", "threads", threads, "max", max(r.threads, len(r.tasks)))

	// Start the specified number of workers.
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() error {
			defer wg.Done()
			for c := range tasks {
				if r.dryRun {
					slog.Debug("DRY-RUN: Run", "cmd", c.String())
				} else {
					if err := c.Run(); err != nil {
						return err
					}
				}
			}
			return nil
		}()
	}

	// When workers are done, close results so that main will exit.
	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		slog.Debug("Print", "result", res)
	}

	return <-results
}
