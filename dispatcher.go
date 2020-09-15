package stop_dispatcher

import (
	"context"

	local_error "github.com/gol4ng/stop-dispatcher/error"
)

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
	stopChan      chan Reason
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

// RegisterCallback is used to register all the wanted stopping callback
func (t *Dispatcher) RegisterCallback(stopCallbacks ...Callback) {
	t.stopCallbacks = append(t.stopCallbacks, stopCallbacks...)
}

// Wait will block until a stopping reason raised from emitter or direct Stop method calling
func (t *Dispatcher) Wait(ctx context.Context) error {
	stopReason := <-t.stopChan
	t.reasonHandler(stopReason)
	shutdownCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := local_error.List{}
	for _, fn := range t.stopCallbacks {
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
