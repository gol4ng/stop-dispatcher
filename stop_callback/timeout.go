package stop_callback

import (
	"context"
	"log"
	"os"
	"time"

	stop_dispatcher "github.com/gol4ng/stop-dispatcher"
)

// local copy for testing purpose
var osExit = os.Exit

// Timeout will exit if the timeout exceed
func Timeout(timeout time.Duration) stop_dispatcher.Callback {
	return func(ctx context.Context) error {
		time.AfterFunc(timeout, func() {
			log.Printf("Shutdown timeout exceeded %s", timeout)
			osExit(1)
		})
		return nil
	}
}
