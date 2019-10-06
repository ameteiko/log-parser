package processor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ameteiko/log-parser/internal/parser"
)

const (
	processingTimeout = 100 * time.Millisecond
)

type Processor struct {
	input         *bufio.Scanner
	stats         statsUpdater
	pendingTraces sync.WaitGroup
	traceLocks    sync.Map
	traceMessages map[string][]*parser.Message
}

type statsUpdater interface {
	IncMalformedLines()
	IncProcessedLines()
	IncOrphanLines()
}

// NewProcessor creates a new Processor object.
func NewProcessor(input *bufio.Scanner, stats statsUpdater) *Processor {
	return &Processor{
		input:         input,
		stats:         stats,
		pendingTraces: sync.WaitGroup{},
		traceLocks:    sync.Map{},
		traceMessages: make(map[string][]*parser.Message),
	}
}

// accumulateMessage accumulates messages to the synchronous map for the postponed processing.
// Using a lock per trace ID to provide concurrency-safety.
func (p *Processor) accumulateMessage(ctx context.Context, msgChan <-chan *parser.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgChan:
			l, ok := p.traceLocks.Load(msg.Trace)
			if !ok {
				p.stats.IncOrphanLines()
				continue
			}
			traceLock := l.(*sync.Mutex)
			traceLock.Lock()
			{
				p.traceMessages[msg.Trace] = append(p.traceMessages[msg.Trace], msg)
			}
			traceLock.Unlock()
		}
	}
}

// Process iterates over all input entries and processes them into the resulting output.
func (p *Processor) Process() {
	ctx, cancel := context.WithCancel(context.Background())
	accumulatorChan := make(chan *parser.Message)

	// TODO: Here is a place for different scaling strategies: we can use multiple accumulator instances here.
	go p.accumulateMessage(ctx, accumulatorChan)

	for p.input.Scan() {
		message, err := parser.Parse(p.input.Text())
		if err != nil {
			p.stats.IncMalformedLines()
			continue
		}

		p.registerTrace(ctx, message.Trace)
		accumulatorChan <- message
		p.stats.IncProcessedLines()
	}

	// Calling cancel signals to all postponed trace processors to generate the result.
	cancel()
	// Await until all processing goroutines finish.
	p.pendingTraces.Wait()
}

// processAfterContextExpiration is a delayed trace ID processor.
// Processing needs to be postponed because the entries in the log file are in a random order, so processing timeout
// ensures that some earlier entries will be collected.
func (p *Processor) processAfterContextExpiration(ctx context.Context, trace string) {
	p.pendingTraces.Add(1)
	defer p.unregisterTrace(trace)
	defer p.pendingTraces.Done()

	<-ctx.Done()

	l, ok := p.traceLocks.Load(trace)
	if !ok {
		// TODO: synchronization problem. Log it here.
		return
	}
	// No lock release because the processing will be over.
	l.(*sync.Mutex).Lock()

	root := findRoot(p.traceMessages[trace])
	if root == nil {
		// TODO: Accumulate statistics on malformed traces.
		return
	}

	result := traceTree{
		ID: trace,
		Root: &call{
			Start:   root.Start,
			End:     root.End,
			Service: root.Service,
			Span:    root.SpanTo,
			Calls:   make([]*call, 0),
		},
	}
	buildTrace(p.traceMessages[trace], result.Root, root.SpanTo)

	res, err := json.Marshal(result)
	if err != nil {
		// TODO: Log an error and update the statistics.
	}

	// TODO: output this result to a proper source destination.
	fmt.Fprintln(os.Stdout, string(res))
}

// registerTrace registers trace as waiting for a processing. Here it means that the mutex is created for it.
// TODO: it would be better to extend the processingTimeout on each new entry.
func (p *Processor) registerTrace(ctx context.Context, trace string) {
	if _, ok := p.traceLocks.Load(trace); ok {
		return
	}

	timeoutCtx, _ := context.WithTimeout(ctx, processingTimeout)
	go p.processAfterContextExpiration(timeoutCtx, trace)

	p.traceLocks.Store(trace, &sync.Mutex{})
}

// registerTrace registers trace as waiting for a processing. Here it means that the mutex is created for it.
func (p *Processor) unregisterTrace(trace string) {
	p.traceLocks.Delete(trace)
}
