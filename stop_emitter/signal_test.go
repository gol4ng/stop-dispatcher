package stop_emitter

import (
	"os"
	"strconv"
	"sync"
	"syscall"
	"testing"

	"github.com/gol4ng/stop-dispatcher"
	"github.com/stretchr/testify/assert"
)

type FakeSignal int

func (s FakeSignal) Signal() {}

func (s FakeSignal) String() string {
	return strconv.Itoa(int(s))
}

func TestDefaultSignalEmitter(t *testing.T) {
	stopCalled := NewSafeBool(false)
	signalNotifyCalled := NewSafeBool(false)
	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled.Set(true)
		assert.Equal(t, []os.Signal{syscall.SIGINT, syscall.SIGTERM}, sig)
		c <- FakeSignal(1)
	}

	fn := DefaultSignalEmitter()
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled.Set(true)
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, stopCalled.Get())
	assert.True(t, signalNotifyCalled.Get())
}

func TestSignalEmitter(t *testing.T) {
	stopCalled := NewSafeBool(false)
	signalNotifyCalled := NewSafeBool(false)
	subscribedSignals := []os.Signal{FakeSignal(1), FakeSignal(2)}

	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled.Set(true)
		assert.Equal(t, subscribedSignals, sig)
		c <- FakeSignal(1)
	}

	fn := SignalEmitter(subscribedSignals...)
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled.Set(true)
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, stopCalled.Get())
	assert.True(t, signalNotifyCalled.Get())
}

func TestDefaultKillerSignalEmitter(t *testing.T) {
	osExitCalled := NewSafeBool(false)
	stopCalled := NewSafeBool(false)
	signalNotifyCalled := NewSafeBool(false)

	// replace local osExit function
	osExit = func(code int) {
		osExitCalled.Set(true)
		assert.Equal(t, 1, code)
	}
	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled.Set(true)
		assert.Equal(t, []os.Signal{syscall.SIGINT, syscall.SIGTERM}, sig)
		c <- FakeSignal(1)
		c <- FakeSignal(1)
	}

	fn := DefaultKillerSignalEmitter()
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled.Set(true)
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, osExitCalled.Get())
	assert.True(t, stopCalled.Get())
	assert.True(t, signalNotifyCalled.Get())
}

func TestKillerSignalEmitter(t *testing.T) {
	osExitCalled := NewSafeBool(false)
	stopCalled := NewSafeBool(false)
	signalNotifyCalled := NewSafeBool(false)
	subscribedSignals := []os.Signal{FakeSignal(1), FakeSignal(2)}

	// replace local signalNotify function
	osExit = func(code int) {
		osExitCalled.Set(true)
		assert.Equal(t, 1, code)
	}
	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled.Set(true)
		assert.Equal(t, subscribedSignals, sig)
		assert.False(t, osExitCalled.Get())
		c <- FakeSignal(1)
		c <- FakeSignal(1)
	}

	fn := KillerSignalEmitter(subscribedSignals...)
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled.Set(true)
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, osExitCalled.Get())
	assert.True(t, stopCalled.Get())
	assert.True(t, signalNotifyCalled.Get())
}

type SafeBool struct {
	b  bool
	mu sync.RWMutex
}

func (s *SafeBool) Get() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.b
}
func (s *SafeBool) Set(b bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.b = b
}
func NewSafeBool(b bool) *SafeBool {
	return &SafeBool{
		b:  b,
		mu: sync.RWMutex{},
	}
}
