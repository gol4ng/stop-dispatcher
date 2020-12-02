package stop_dispatcher

import (
	"context"
	"sync"

	local_error "github.com/gol4ng/stop-dispatcher/error"
)

// use to replace the callback when they unregistered
var nopFunc = func(ctx context.Context) error { return nil }

// Reason is the stopping given value
type Reason interface{}

// Emitter can emit a reason that will be dispatched
type Emitter func(func(Reason))

// It receive the emitted stop reason before calling callbacks
type ReasonHandler func(Reason)

// Callback will be called when a reason raised from Emitter
type Callback func(ctx context.Context) error

// Dispatcher implementation provide Reason dispatcher
type Dispatcher struct {
	stopChan chan Reason

	mu            sync.RWMutex
	stopCallbacks []Callback

	reasonHandler ReasonHandler
}

// Stop is the begin of stopping dispatch
func (t *Dispatcher) Stop(reason Reason) {
	t.stopChan <- reason
}

// RegisterEmitter is used to register all the wanted emitter
func (t *Dispatcher) RegisterEmitter(stopEmitters ...Emitter) {
	for _, stopEmitter := range stopEmitters {
		go stopEmitter(t.Stop)
	}
}

// RegisterCallback is used to register stopping callback
// It return a func to unregister the callback
func (t *Dispatcher) RegisterCallback(stopCallback Callback) func() {
	i := len(t.stopCallbacks)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopCallbacks = append(t.stopCallbacks, stopCallback)

	return func() {
		t.mu.Lock()
		defer t.mu.Unlock()
		t.stopCallbacks[i] = nopFunc
	}
}

// RegisterCallbacks is used to register all the wanted stopping callback
// With this method you cannot unregister a callback
// If you want to unregister callback you should use RegisterCallback
func (t *Dispatcher) RegisterCallbacks(stopCallbacks ...Callback) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopCallbacks = append(t.stopCallbacks, stopCallbacks...)
}

// Wait will block until a stopping reason raised from emitter or direct Stop method calling
func (t *Dispatcher) Wait(ctx context.Context) error {
	stopReason := <-t.stopChan
	t.reasonHandler(stopReason)
	shutdownCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := local_error.List{}
	t.mu.RLock()
	stopCallbacks := t.stopCallbacks
	defer t.mu.RUnlock()
	for _, fn := range stopCallbacks {
		if err := fn(shutdownCtx); err != nil {
			errs.Add(err)
		}
	}
	return errs.ReturnOrNil()
}

// NewDispatcher construct a new Dispatcher with the given options
func NewDispatcher(options ...DispatcherOption) *Dispatcher {
	dispatcher := &Dispatcher{
		stopChan:      make(chan Reason),
		mu:            sync.RWMutex{},
		stopCallbacks: []Callback{},
		reasonHandler: func(Reason) {},
	}

	for _, option := range options {
		option(dispatcher)
	}

	return dispatcher
}

// DispatcherOption represent a Dispatcher option
type DispatcherOption func(*Dispatcher)

// WithReasonHandler configure a reason handler
func WithReasonHandler(reasonHandler ReasonHandler) DispatcherOption {
	return func(dispatcher *Dispatcher) {
		dispatcher.reasonHandler = reasonHandler
	}
}

// WithEmitter is a helpers to register during the construction
func WithEmitter(emitters ...Emitter) DispatcherOption {
	return func(dispatcher *Dispatcher) {
		dispatcher.RegisterEmitter(emitters...)
	}
}
