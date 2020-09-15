package stop_dispatcher_test

import (
	"context"
	"errors"
	"testing"
	"time"

	stop_dispatcher "github.com/gol4ng/stop-dispatcher"
	"github.com/stretchr/testify/assert"
)

func Test_Dispatcher_Stop(t *testing.T) {
	callbackCalled := false
	d := stop_dispatcher.NewDispatcher()
	d.RegisterCallback(func(ctx context.Context) error {
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
	callbackCalled := false
	d := stop_dispatcher.NewDispatcher()
	d.RegisterCallback(
		func(ctx context.Context) error {
			callbackCalled = true
			return errors.New("fake_error")
		},
		func(ctx context.Context) error {
			return errors.New("fake_error2")
		},
	)
	time.AfterFunc(10*time.Millisecond, func() {
		d.Stop("fake_reason")
	})

	err := d.Wait(context.TODO())
	assert.EqualError(t, err, "fake_error\nfake_error2\n")
	assert.True(t, callbackCalled)
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

func Test_Dispatcher(t *testing.T) {
	callbackCalled := false
	d := stop_dispatcher.NewDispatcher()
	d.RegisterCallback(func(ctx context.Context) error {
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