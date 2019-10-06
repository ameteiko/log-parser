package parser

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/xerrors"
)

const (
	datetimeRegExp = `[\dTZ\.\-\:]*`
	traceRegExp    = `[\w-]{1,}`
)

var (
	// Validation errors.
	errLogMessageParsing     = errors.New("log message parsing error")
	errLogMessageDateInvalid = errors.New("date is invalid in log message")
)

// Parse parses a log line into a Message object.
func Parse(msg string) (*Message, error) {
	// TODO: improve regular expression rules: use word boundaries instead spaces, improve UTC timestamp regular
	//  expression, etc.
	logMessagePattern := fmt.Sprintf("(%[1]s) (%[1]s) (%[2]s) (%[2]s) (%[2]s)->(%[2]s)", datetimeRegExp, traceRegExp)
	logMessageRE := regexp.MustCompile(logMessagePattern)

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
