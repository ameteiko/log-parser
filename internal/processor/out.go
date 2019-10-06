package processor

import (
	"encoding/json"
	"time"
)

// traceTree is a resulting trace tree for the traces.
type traceTree struct {
	ID   string `json:"id"`
	Root *call  `json:"root"`
}

// call represents a an output call entry.
// The parser.Message type is not reused for the sake of separation of concerns.
type call struct {
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Service string    `json:"service"`
	Span    string    `json:"span"`
	Calls   []*call   `json:"calls"`
}

// MarshalJSON overrides marshaling of the time fields.
func (d *call) MarshalJSON() ([]byte, error) {
	type alias call
	return json.Marshal(&struct {
		Start string `json:"start"`
		End   string `json:"end"`
		*alias
	}{
		Start: d.Start.Format(time.RFC3339Nano),
		End:   d.End.Format(time.RFC3339Nano),
		alias: (*alias)(d),
	})
}
