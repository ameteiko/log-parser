package processor

import (
	"bufio"
	"context"
	"sync"
	"time"

	"github.com/ameteiko/log-parser/internal/parser"
)

const (
	processingTimeout = 100 * time.Millisecond
)

type Processor struct {
	input           *bufio.Scanner
	stats           statsUpdater
	pendingTraces   sync.WaitGroup
	pendingTraceIDs sync.Map
	traceMessages   map[string][]*parser.Message
}

type statsUpdater interface {
	IncMalformedLines()
	IncProcessedLines()
	IncOrphanLines()
}

// NewProcessor creates a new Processor object.
func NewProcessor(input *bufio.Scanner, stats statsUpdater) *Processor {
	return &Processor{
		input:           input,
		stats:           stats,
		pendingTraces:   sync.WaitGroup{},
		pendingTraceIDs: sync.Map{},
		traceMessages:   make(map[string][]*parser.Message),
	}
}

// Process iterates over all input entries processing them into the resulting output.
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
	cancel()
	p.pendingTraces.Wait()
}

// processAfterContextExpiration is a delayed trace ID processor.
// Processing needs to be postponed because the entries in the log file are in a random order, so processing timeout
// ensures that some earlier entries will be collected.
func (p *Processor) processAfterContextExpiration(ctx context.Context, trace string) {
	p.pendingTraces.Add(1)

	<-ctx.Done()
	println("postprocessing ", trace)
	// Build a priority queue

	p.unregisterTrace(trace)
	p.pendingTraces.Done()
}

// registerTrace registers trace as waiting for a processing. Here it means that the mutex is created for it.
func (p *Processor) registerTrace(ctx context.Context, trace string) {
	if _, ok := p.pendingTraceIDs.Load(trace); ok {
		return
	}

	timeoutCtx, _ := context.WithTimeout(ctx, processingTimeout)
	go p.processAfterContextExpiration(timeoutCtx, trace)

	p.pendingTraceIDs.Store(trace, &sync.Mutex{})
}

// registerTrace registers trace as waiting for a processing. Here it means that the mutex is created for it.
func (p *Processor) unregisterTrace(trace string) {
	p.pendingTraceIDs.Delete(trace)
}

// accumulateMessage accumulates messages to the synchronous map for the postponed processing.
// I use it as a goroutine to make ready for scalability.
func (p *Processor) accumulateMessage(ctx context.Context, msgChan <-chan *parser.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgChan:
			l, ok := p.pendingTraceIDs.Load(msg.Trace)
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
