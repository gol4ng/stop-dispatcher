package reason_handler

import (
	"log"
	"os"

	stop_dispatcher "github.com/gol4ng/stop-dispatcher"
)

// Log was a default log reason handler
func Log() stop_dispatcher.ReasonHandler {
	return func(reason stop_dispatcher.Reason) {
		switch value := reason.(type) {
		case os.Signal:
			log.Printf("received signal (%s)", value.String())
		case error:
			log.Printf("fatal error : %s", value.Error())
		default:
			log.Printf("stop reason %v", value)
		}
	}
}
