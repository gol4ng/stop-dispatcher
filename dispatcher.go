package stop_dispatcher

import (
	"context"
	"sort"
	"sync"

	local_error "github.com/gol4ng/stop-dispatcher/error"
)

// use to replace the callback when they unregistered
var nopFunc = CallbackFunc(func(ctx context.Context) error { return nil })

// Reason is the stopping given value
type Reason interface{}

// Emitter can emit a reason that will be dispatched
type Emitter func(func(Reason))

// It receive the emitted stop reason before calling callbacks
type ReasonHandler func(Reason)

// CallbackFunc will be called when a reason raised from Emitter
type CallbackFunc func(ctx context.Context) error

// GetPriority will return callback priority
func (c CallbackFunc) GetPriority() int {
	return 0
}

// Callback will execute the CallbackFunc
func (c CallbackFunc) Callback(ctx context.Context) error {
	return c(ctx)
}

// Callback represent a standard callback
type Callback interface {
	GetPriority() int
	Callback(ctx context.Context) error
}

// PrioritizedCallback represent a callback with priority defined
type PrioritizedCallback struct {
	CallbackFunc
	priority int
}

// GetPriority will return callback priority
func (c PrioritizedCallback) GetPriority() int {
	return c.priority
}

// NewPrioritizeCallback wrap CallbackFunc to configure custom priority
func NewPrioritizeCallback(priority int, callback CallbackFunc) Callback {
	return PrioritizedCallback{
		CallbackFunc: callback,
		priority:     priority,
	}
}

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

// RegisterPrioritizeCallbackFunc will register a CallbackFunc with the given priority
// It return a func to unregister the callback
func (t *Dispatcher) RegisterPrioritizeCallbackFunc(priority int, stopCallback CallbackFunc) func() {
	return t.RegisterCallback(NewPrioritizeCallback(priority, stopCallback))
}

// RegisterCallbackFunc will register a CallbackFunc with the priority at 0
// It return a func to unregister the callback
func (t *Dispatcher) RegisterCallbackFunc(stopCallback CallbackFunc) func() {
	return t.RegisterCallback(stopCallback)
}

// RegisterCallbacksFunc will register multiple CallbackFunc with the priority at 0
// If you want to unregister callback you should use RegisterCallback
func (t *Dispatcher) RegisterCallbacksFunc(stopCallbacks ...CallbackFunc) {
	callbacks := make([]Callback, len(stopCallbacks))
	for i, c := range stopCallbacks {
		callbacks[i] = c
	}
	t.RegisterCallbacks(callbacks...)
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
	t.mu.RUnlock()
	// Sort stopCallbacks highest priority first
	sort.Slice(stopCallbacks, func(i, j int) bool {
		return stopCallbacks[i].GetPriority() > stopCallbacks[j].GetPriority()
	})
	for _, fn := range stopCallbacks {
		if err := fn.Callback(shutdownCtx); err != nil {
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
