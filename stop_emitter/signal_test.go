package stop_emitter

import (
	"os"
	"strconv"
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
	stopCalled := false
	signalNotifyCalled := false
	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled = true
		assert.Equal(t, []os.Signal{syscall.SIGINT, syscall.SIGTERM}, sig)
		c <- FakeSignal(1)
	}

	fn := DefaultSignalEmitter()
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled = true
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, stopCalled)
	assert.True(t, signalNotifyCalled)
}

func TestSignalEmitter(t *testing.T) {
	stopCalled := false
	signalNotifyCalled := false
	subscribedSignals := []os.Signal{FakeSignal(1), FakeSignal(2)}

	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled = true
		assert.Equal(t, subscribedSignals, sig)
		c <- FakeSignal(1)
	}

	fn := SignalEmitter(subscribedSignals...)
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled = true
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, stopCalled)
	assert.True(t, signalNotifyCalled)
}

func TestDefaultKillerSignalEmitter(t *testing.T) {
	osExitCalled := false
	stopCalled := false
	signalNotifyCalled := false

	// replace local osExit function
	osExit = func(code int) {
		osExitCalled = true
		assert.Equal(t, 1, code)
	}
	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled = true
		assert.Equal(t, []os.Signal{syscall.SIGINT, syscall.SIGTERM}, sig)
		c <- FakeSignal(1)
		c <- FakeSignal(1)
	}

	fn := DefaultKillerSignalEmitter()
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled = true
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, osExitCalled)
	assert.True(t, stopCalled)
	assert.True(t, signalNotifyCalled)
}

func TestKillerSignalEmitter(t *testing.T) {
	osExitCalled := false
	stopCalled := false
	signalNotifyCalled := false
	subscribedSignals := []os.Signal{FakeSignal(1), FakeSignal(2)}

	// replace local signalNotify function
	osExit = func(code int) {
		osExitCalled = true
		assert.Equal(t, 1, code)
	}
	// replace local signalNotify function
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		signalNotifyCalled = true
		assert.Equal(t, subscribedSignals, sig)
		assert.False(t, osExitCalled)
		c <- FakeSignal(1)
		c <- FakeSignal(1)
	}

	fn := KillerSignalEmitter(subscribedSignals...)
	wait := make(chan struct{})
	go fn(func(reason stop_dispatcher.Reason) {
		stopCalled = true
		assert.Equal(t, FakeSignal(1), reason)
		close(wait)
	})
	<-wait
	assert.True(t, osExitCalled)
	assert.True(t, stopCalled)
	assert.True(t, signalNotifyCalled)
}
