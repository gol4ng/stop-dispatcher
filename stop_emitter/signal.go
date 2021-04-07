package stop_emitter

import (
	"os"
	"os/signal"
	"syscall"

	stop_dispatcher "github.com/gol4ng/stop-dispatcher"
)

// local copy for testing purpose
var osExit = os.Exit
var signalNotify = signal.Notify

// DefaultSignalEmitter will emit stop reason when syscall.SIGINT, syscall.SIGTERM was received
func DefaultSignalEmitter() stop_dispatcher.Emitter {
	return SignalEmitter(syscall.SIGINT, syscall.SIGTERM)
}

// SignalEmitter will emit stop reason when signal was received
func SignalEmitter(signals ...os.Signal) stop_dispatcher.Emitter {
	return func(stop stop_dispatcher.ReasonHandler) {
		signalChan := make(chan os.Signal, 1)
		signalNotify(signalChan, signals...)
		stop(<-signalChan)
	}
}

// DefaultKillerSignalEmitter will emit stop reason when syscall.SIGINT, syscall.SIGTERM was received
// it exit if signal was received a second time
func DefaultKillerSignalEmitter() stop_dispatcher.Emitter {
	return KillerSignalEmitter(syscall.SIGINT, syscall.SIGTERM)
}

// KillerSignalEmitter will emit stop reason when signal was received
// it exit if signal was received a second time
func KillerSignalEmitter(signals ...os.Signal) stop_dispatcher.Emitter {
	return func(stop stop_dispatcher.ReasonHandler) {
		signalReceived := false
		signalChan := make(chan os.Signal, 2)
		signalNotify(signalChan, signals...)
		for reason := range signalChan {
			if !signalReceived {
				signalReceived = true
				stop(reason)
			} else {
				osExit(1)
			}
		}
	}
}
