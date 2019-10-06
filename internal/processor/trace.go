package processor

import (
	"sort"

	"github.com/ameteiko/log-parser/internal/parser"
)

const (
	rootSpanID = "null"
)

// findRoot returns a root trace message.
// Current implementation ignores the possibility of having several roots for the same trace ID.
func findRoot(messages []*parser.Message) *parser.Message {
	for _, msg := range messages {
		if msg.SpanFrom == rootSpanID {
			return msg
		}
	}

	return nil
}

// buildTrace is a brute-force implementation for the building the call-tree.
// TODO: replace with possible implementation based on building a priority queue for the messages (by the msg.Start),
//  then iterate over that queue building a resulting calls tree (such implementation will allow to identify orphan
//  entries).
func buildTrace(messages []*parser.Message, callTrace *call, spanTo string) {
	var matches []*parser.Message
	for _, msg := range messages {
		if msg.SpanFrom == spanTo {
			matches = append(matches, msg)
		}
	}
	sort.Sort(parser.Messages(matches))

	for _, m := range matches {
		c := call{
			Start:   m.Start,
			End:     m.End,
			Service: m.Service,
			Span:    m.SpanTo,
			Calls:   make([]*call, 0),
		}
		buildTrace(messages, &c, m.SpanTo)
		callTrace.Calls = append(callTrace.Calls, &c)
	}
}
