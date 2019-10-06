package stats

import (
	"sync"
)

// Stats collects log processing statistics.
type Stats struct {
	mx             sync.Mutex
	malformedLines int
	processedLines int
	orphanLines    int
}

// NewStats creates a new Stats object.
// TODO: use separate locks and use read/write locks for different statistics parameters.
func NewStats() *Stats {
	return &Stats{mx: sync.Mutex{}}
}

// IncMalformedLines increments the number of the erroneous log lines.
func (s *Stats) IncMalformedLines() {
	s.mx.Lock()
	s.malformedLines++
	s.mx.Unlock()
}

// IncProcessedLines increments the number of the processed lines.
func (s *Stats) IncProcessedLines() {
	s.mx.Lock()
	s.processedLines++
	s.mx.Unlock()
}

// IncOrphanLines increments the number of the orphan lines: the ones that are not related to any log trace.
func (s *Stats) IncOrphanLines() {
	s.mx.Lock()
	s.orphanLines++
	s.mx.Unlock()
}
