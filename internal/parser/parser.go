package parser

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/xerrors"
)

const (
	reDatetime = `[\dTZ\.\-\:]*`
	reTrace    = `[\w-]{1,}`
)

var (
	// TODO: improve regular expression rules: use word boundaries instead spaces, improve UTC timestamp regular
	//  expression, etc.
	logMessageRE = regexp.MustCompile(fmt.Sprintf("(%[1]s) (%[1]s) (%[2]s) (%[2]s) (%[2]s)->(%[2]s)", reDatetime, reTrace))

	// Validation errors.
	errLogMessageParsing     = errors.New("log message parsing error")
	errLogMessageDateInvalid = errors.New("date is invalid in log message")
)

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

// Parse parses a log line into a Message object.
func Parse(msg string) (*Message, error) {
	fields := logMessageRE.FindStringSubmatch(msg)
	if len(fields) == 0 {
		return nil, errLogMessageParsing
	}

	// TODO: add additional fine-grained validation for the non-timestamp fields.
	start := fields[1]
	end := fields[2]
	trace := fields[3]
	service := fields[4]
	spanFrom := fields[5]
	spanTo := fields[6]

	startDatetime, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		return nil, xerrors.Errorf("log message start date %s is invalid: %v", fields[1], errLogMessageDateInvalid)
	}

	endDatetime, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		return nil, xerrors.Errorf("log message end date %s is invalid: %v", fields[2], errLogMessageDateInvalid)
	}

	return &Message{
		Service:  service,
		Trace:    trace,
		SpanFrom: spanFrom,
		SpanTo:   spanTo,
		Start:    startDatetime,
		End:      endDatetime,
	}, nil
}
