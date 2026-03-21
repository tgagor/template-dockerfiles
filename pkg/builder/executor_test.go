package builder_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tgagor/template-dockerfiles/pkg/builder"
	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/image"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
)

type mockBuilder struct {
	mu            sync.Mutex
	processCalls  []string
	processDelays map[string]time.Duration
	startTimes    map[string]time.Time
	endTimes      map[string]time.Time
}

func newMockBuilder() *mockBuilder {
	return &mockBuilder{
		processDelays: make(map[string]time.Duration),
		startTimes:    make(map[string]time.Time),
		endTimes:      make(map[string]time.Time),
	}
}

func (m *mockBuilder) Init() error                  { return nil }
func (m *mockBuilder) SetFlags(flags *config.Flags) {}
func (m *mockBuilder) Terminate() error             { return nil }
func (m *mockBuilder) Process(ctx context.Context, img *image.Image) error {
	m.mu.Lock()
	m.startTimes[img.Name] = time.Now()
	m.processCalls = append(m.processCalls, img.Name)
	delay := m.processDelays[img.Name]
	m.mu.Unlock()

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	m.mu.Lock()
	m.endTimes[img.Name] = time.Now()
	m.mu.Unlock()
	return nil
}

func TestExecutePlan_DAGConcurrency(t *testing.T) {
	// Plan structure:
	// A -> B -> D
	// A -> C -> D
	// E (independent)

	plan := &parser.Plan{
		Nodes: map[string]*parser.Node{
			"A": {ID: "A", Image: &image.Image{Name: "A"}, DependsOn: []string{}},
			"B": {ID: "B", Image: &image.Image{Name: "B"}, DependsOn: []string{"A"}},
			"C": {ID: "C", Image: &image.Image{Name: "C"}, DependsOn: []string{"A"}},
			"D": {ID: "D", Image: &image.Image{Name: "D"}, DependsOn: []string{"B", "C"}},
			"E": {ID: "E", Image: &image.Image{Name: "E"}, DependsOn: []string{}},
		},
	}

	mock := newMockBuilder()
	// A takes 50ms, B takes 50ms, C takes 150ms, E takes 200ms
	mock.processDelays["A"] = 50 * time.Millisecond
	mock.processDelays["B"] = 50 * time.Millisecond
	mock.processDelays["C"] = 150 * time.Millisecond
	mock.processDelays["D"] = 10 * time.Millisecond
	mock.processDelays["E"] = 200 * time.Millisecond

	flags := &config.Flags{Threads: 4}

	err := builder.ExecutePlan(plan, mock, flags)
	require.NoError(t, err)

	mock.mu.Lock()
	defer mock.mu.Unlock()

	assert.Len(t, mock.processCalls, 5)

	// B must start at or after A ends
	assert.False(t, mock.startTimes["B"].Before(mock.endTimes["A"]))

	// C must start at or after A ends
	assert.False(t, mock.startTimes["C"].Before(mock.endTimes["A"]))

	// D must start at or after B ends and C ends
	assert.False(t, mock.startTimes["D"].Before(mock.endTimes["B"]))
	assert.False(t, mock.startTimes["D"].Before(mock.endTimes["C"]))

	// E should have started concurrently with A
	assert.True(t, mock.startTimes["E"].Before(mock.endTimes["A"]))
}
