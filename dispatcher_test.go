package stop_dispatcher_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	stop_dispatcher "github.com/gol4ng/stop-dispatcher"
	"github.com/stretchr/testify/assert"
)

func Test_Dispatcher_Stop(t *testing.T) {
	callbackCalled := false
	d := stop_dispatcher.NewDispatcher()
	d.RegisterCallbackFunc(func(ctx context.Context) error {
		callbackCalled = true
		return nil
	})
	time.AfterFunc(10*time.Millisecond, func() {
		d.Stop("fake_reason")
	})

	err := d.Wait(context.TODO())
	assert.NoError(t, err)
	assert.True(t, callbackCalled)
}

func Test_Dispatcher_Error(t *testing.T) {
	var step []string
	d := stop_dispatcher.NewDispatcher()
	d.RegisterPrioritizeCallbackFunc(1, func(ctx context.Context) error {
		step = append(step, "priority:1_index:0")
		return errors.New("fake_error_priority_1")
	})
	d.RegisterCallbacksFunc(
		func(ctx context.Context) error {
			step = append(step, "priority:0_index:0")
			return errors.New("fake_error")
		},
		func(ctx context.Context) error {
			step = append(step, "priority:0_index:1")
			return errors.New("fake_error2")
		},
	)
	time.AfterFunc(10*time.Millisecond, func() {
		d.Stop("fake_reason")
	})

	err := d.Wait(context.TODO())
	assert.EqualError(t, err, "fake_error_priority_1\nfake_error\nfake_error2\n")
	assert.Len(t, step, 3)
	assert.Equal(t, []string{
		"priority:1_index:0",
		"priority:0_index:0",
		"priority:0_index:1",
	}, step)
}

func Test_Dispatcher_WithReasonHandler(t *testing.T) {
	reasonHandlerCalled := false
	d := stop_dispatcher.NewDispatcher(
		stop_dispatcher.WithReasonHandler(func(reason stop_dispatcher.Reason) {
			reasonHandlerCalled = true
			assert.Equal(t, "fake_reason", reason)
		}),
	)
	time.AfterFunc(10*time.Millisecond, func() {
		d.Stop("fake_reason")
	})

	err := d.Wait(context.TODO())
	assert.NoError(t, err)
	assert.True(t, reasonHandlerCalled)
}

func Test_Dispatcher_WithEmitter(t *testing.T) {
	d := stop_dispatcher.NewDispatcher(
		stop_dispatcher.WithEmitter(func(stopFn func(stop_dispatcher.Reason)) {
			time.AfterFunc(10*time.Millisecond, func() {
				stopFn("fake_reason")
			})
		}),
	)

	err := d.Wait(context.TODO())
	assert.NoError(t, err)
}

func Test_Dispatcher_UnregisterCallback(t *testing.T) {
	safeInnerStopFn := sync.Mutex{}
	var innerStopFn func(stop_dispatcher.Reason)
	d := stop_dispatcher.NewDispatcher(
		stop_dispatcher.WithEmitter(func(stopFn func(stop_dispatcher.Reason)) {
			safeInnerStopFn.Lock()
			innerStopFn = stopFn
			safeInnerStopFn.Unlock()
		}),
	)
	callbackCalled := false
	unregisterCallbackFunc := d.RegisterCallbackFunc(func(ctx context.Context) error {
		callbackCalled = true
		return nil
	})
	go func() {
		time.AfterFunc(10*time.Millisecond, func() {
			safeInnerStopFn.Lock()
			innerStopFn("fake_reason")
			safeInnerStopFn.Unlock()
		})
	}()
	err := d.Wait(context.TODO())
	assert.NoError(t, err)
	assert.True(t, callbackCalled)

	callbackCalled = false
	unregisterCallbackFunc()
	go func() {
		time.AfterFunc(10*time.Millisecond, func() {
			safeInnerStopFn.Lock()
			innerStopFn("fake_reason")
			safeInnerStopFn.Unlock()
		})
	}()
	err1 := d.Wait(context.TODO())
	assert.NoError(t, err1)
	assert.False(t, callbackCalled)
}

func Test_Dispatcher(t *testing.T) {
	callbackCalled := false
	d := stop_dispatcher.NewDispatcher()
	d.RegisterCallbackFunc(func(ctx context.Context) error {
		callbackCalled = true
		return nil
	})
	d.RegisterEmitter(func(stopFn func(stop_dispatcher.Reason)) {
		time.AfterFunc(10*time.Millisecond, func() {
			stopFn("fake_reason")
		})
	})

	err := d.Wait(context.TODO())
	assert.NoError(t, err)
	assert.True(t, callbackCalled)
}

func Test_Dispatcher_Prioritize(t *testing.T) {
	var step []string
	d := stop_dispatcher.NewDispatcher()
	d.RegisterCallbacksFunc(
		func(ctx context.Context) error {
			step = append(step, "priority:0_index:0")
			return nil
		},
		func(ctx context.Context) error {
			step = append(step, "priority:0_index:1")
			return nil
		},
	)
	d.RegisterCallbacks(
		stop_dispatcher.NewPrioritizeCallback(1, func(ctx context.Context) error {
			step = append(step, "priority:1_index:0")
			return nil
		}),
		stop_dispatcher.NewPrioritizeCallback(0, func(ctx context.Context) error {
			step = append(step, "priority:0_index:2")
			return nil
		}),
	)
	d.RegisterPrioritizeCallbackFunc(3, func(ctx context.Context) error {
		step = append(step, "priority:3_index:0")
		return nil
	})
	d.RegisterEmitter(func(stopFn func(stop_dispatcher.Reason)) {
		time.AfterFunc(10*time.Millisecond, func() {
			stopFn("fake_reason")
		})
	})
	err := d.Wait(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, step, 5)
	assert.Equal(t, []string{
		"priority:3_index:0",
		"priority:1_index:0",
		"priority:0_index:0",
		"priority:0_index:1",
		"priority:0_index:2",
	}, step)
}
