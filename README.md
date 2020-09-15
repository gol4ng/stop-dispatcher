# Stop dispatcher

## Installation

`go get -u github.com/gol4ng/stop-dispatcher`

```go
package main

import (
	"context"
	"log"

	stop_dispatcher "github.com/gol4ng/stop-dispatcher"
	"github.com/gol4ng/stop-dispatcher/reason_handler"
	"github.com/gol4ng/stop-dispatcher/stop_emitter"
)

func main() {
	// root context
	ctx := context.Background()

	// stop dispatcher with log reason handler
	stopDispatcher := stop_dispatcher.NewDispatcher(
		stop_dispatcher.WithReasonHandler(reason_handler.Log()),
	)

	// Register a killer signal emitter
	stopDispatcher.RegisterEmitter(
		stop_emitter.DefaultKillerSignalEmitter(),
	)

	// Register all your stopping callback
	stopDispatcher.RegisterCallback(
		func(ctx context.Context) error {
			log.Println("Closing all database connection")
			return nil
		},
		func(ctx context.Context) error {
			log.Println("Closing all server connection")
			return nil
		},
	)

	// Wait will block until stopping reason was received
	if err := stopDispatcher.Wait(ctx); err != nil {
		log.Printf("error occured during stopping application : %s", err)
	}
	log.Println("Application stopped")
}
```