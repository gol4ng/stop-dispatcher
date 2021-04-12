package stop_callback

import (
	"bytes"
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimedout(t *testing.T) {
	// configure log to capture data
	buf := SafeWrapBuffer(bytes.NewBuffer([]byte{}))
	log.SetFlags(0)
	log.SetOutput(buf)

	// replace local osexit function
	osExit = func(code int) {
		assert.Equal(t, 1, code)
	}

	fn := Timeout(100 * time.Millisecond)
	assert.Nil(t, fn.Callback(context.TODO()))
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, "Shutdown timeout exceeded 100ms\n", buf.String())
}

type Buffer struct {
	b *bytes.Buffer
	m sync.Mutex
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}
func (b *Buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func SafeWrapBuffer(buffer *bytes.Buffer) *Buffer {
	return &Buffer{
		b: buffer,
		m: sync.Mutex{},
	}
}
