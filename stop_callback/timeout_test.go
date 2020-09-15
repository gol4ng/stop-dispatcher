package stop_callback

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimedout(t *testing.T) {
	// configure log to capture data
	buf := bytes.NewBuffer([]byte{})
	log.SetFlags(0)
	log.SetOutput(buf)

	// replace local osexit function
	osExit = func(code int) {
		assert.Equal(t, 1, code)
	}

	fn := Timeout(100 * time.Millisecond)
	assert.Nil(t, fn(context.TODO()))
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, "Shutdown timeout exceeded 100ms\n", buf.String())
}
