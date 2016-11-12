
package fahclient

import (
  "log"
)

func logFatalUnknownErr(errType string, err error) {
  log.Fatalf("Unknown %v (%T): %v\n", errType, err, err)
}
