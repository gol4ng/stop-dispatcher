package reason_handler_test

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/gol4ng/stop-dispatcher/reason_handler"
	"github.com/stretchr/testify/assert"
)

type FakeSignal int

func (s FakeSignal) Signal() {}

func (s FakeSignal) String() string {
	return strconv.Itoa(int(s))
}

func TestLog(t *testing.T) {
	// configure log to capture data
	buf := bytes.NewBuffer([]byte{})
	log.SetOutput(buf)
	log.SetFlags(0)

	reasonHandler := reason_handler.Log()

	tests := []struct {
		input          interface{}
		expectedOutput string
	}{
		// signal
		{input: FakeSignal(1), expectedOutput: "received signal (1)\n"},

		// error
		{input: errors.New("fake_error"), expectedOutput: "fatal error : fake_error\n"},

		// default
		{input: nil, expectedOutput: "stop reason <nil>\n"},
		{input: true, expectedOutput: "stop reason true\n"},
		{input: 1, expectedOutput: "stop reason 1\n"},
		{input: "fake_string", expectedOutput: "stop reason fake_string\n"},
		{input: time.Second, expectedOutput: "stop reason 1s\n"},
		{input: struct{}{}, expectedOutput: "stop reason {}\n"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			reasonHandler(tt.input)
			assert.Equal(t, tt.expectedOutput, buf.String())
			buf.Reset()
		})
	}
}
