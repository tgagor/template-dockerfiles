package builder

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
	"github.com/tgagor/template-dockerfiles/pkg/tui"
)

// ExecutePlan orchestrates the build process across all nodes of the dependency graph,
// processing ready images non-blocking through a worker pool.
func ExecutePlan(plan *parser.Plan, b Builder, flags *config.Flags, events chan<- tui.EventMsg) error {
	if err := b.Init(); err != nil {
		return err
	}
	b.SetFlags(flags)
	defer func() {
		if err := b.Terminate(); err != nil {
			log.Warn().Err(err).Msg("Builder termination failed")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inDegree := make(map[string]int)
	outEdges := make(map[string][]string)

	for id := range plan.Nodes {
		inDegree[id] = 0
	}

	for id, node := range plan.Nodes {
		for _, dep := range node.DependsOn {
			inDegree[id]++
			outEdges[dep] = append(outEdges[dep], id)
		}
	}

	// Buffer size at least len(Nodes) so we never deadlock
	readyCh := make(chan string, len(plan.Nodes))

	for id, degree := range inDegree {
		if degree == 0 {
			readyCh <- id
		}
	}

	var mu sync.Mutex // guards inDegree and processedCount
	var processedCount int

	workerCount := flags.Threads
	if workerCount < 1 {
		workerCount = 1
	}

	log.Info().Int("total_images", len(plan.Nodes)).Int("workers", workerCount).Msg("Starting parallel DAG execution")

	var wg sync.WaitGroup
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case id, ok := <-readyCh:
					if !ok {
						return // channel closed, done
					}

					node := plan.Nodes[id]
					log.Info().Str("image", node.Image.Name).Msg("Processing")

					if err := b.Process(ctx, node.Image, events); err != nil {
						log.Error().Err(err).Str("image", node.Image.Name).Msg("Processing failed")
						if events != nil {
							events <- tui.EventMsg{Err: err}
						}
						errCh <- err
						cancel() // cancel context for other workers
						return
					}

					if events != nil {
						events <- tui.EventMsg{ImageName: node.Image.Name, IsDone: true}
					}

					mu.Lock()
					processedCount++
					done := processedCount == len(plan.Nodes)
					for _, dependent := range outEdges[id] {
						inDegree[dependent]--
						if inDegree[dependent] == 0 {
							readyCh <- dependent
						}
					}
					mu.Unlock()

					if done {
						close(readyCh)
						if events != nil {
							close(events)
						}
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
