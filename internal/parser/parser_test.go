package parser

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tcs := []struct {
		name       string
		logMessage string
		message    *Message
		err        error
	}{
		// TODO: add more tests cases for all kinds of possible errors.
		{
			name:       "ForAnEmptyMessage",
			logMessage: "",
			message:    nil,
			err:        errLogMessageParsing,
		},
		{
			name:       "ForAValidMessage",
			logMessage: "2016-10-20T12:43:34.000Z 2016-10-20T12:43:35.000Z trace1 back-end-3 ac->ad",
			message: &Message{
				Service:  "back-end-3",
				Trace:    "trace1",
				SpanFrom: "ac",
				SpanTo:   "ad",
				Start:    time.Date(2016, 10, 20, 12, 43, 34, 0, time.UTC),
				End:      time.Date(2016, 10, 20, 12, 43, 35, 0, time.UTC),
			},
			err: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := Parse(tc.logMessage)
			// TODO: readjust this comparison when more tests are added.
			if tc.message != nil && msg != nil && *tc.message != *msg {
				t.Errorf("Parse(%q) != %v, got %v", tc.message, tc.message, msg)
			}
			if tc.err != err {
				t.Errorf("Parse(%q) returned unexpected error %v != %v", tc.message, tc.err, err)
			}
		})
	}
}
