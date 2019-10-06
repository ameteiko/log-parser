package parser

import (
	"time"
)

type Messages []*Message

// Message represents log file data entry.
// Messaged with fields 2016-10-20T12:43:34.000Z 2016-10-20T12:43:35.000Z trace1 back-end-3 ac->ad is parsed to:
//
// 2016-10-20T12:43:34.000Z - Start
// 2016-10-20T12:43:35.000Z - End
// trace1                  - Trace
// back-end-3              - Service
// ac                      - SpanFrom
// ad                      - SpanTo
type Message struct {
	Service  string
	Trace    string
	SpanFrom string
	SpanTo   string
	Start    time.Time
	End      time.Time
}

// Len returns the len for the messages collection.
func (m Messages) Len() int {
	return len(m)
}

// Less return true if m[i] is before m[j].
func (m Messages) Less(i, j int) bool {
	return m[i].Start.Before(m[j].Start)
}

// Swap swaps m[i] and m[j].
func (m Messages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
