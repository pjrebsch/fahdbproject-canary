
package main

import (
  "github.com/pjrebsch/fahdbproject-canary/fahclient"
  "os"
  "log"
  "sync"
  "errors"
  "bytes"
  "encoding/hex"
)

// Flag for signaling that the application should exit due to
// either an unrecoverable error or the user's request.
var pleaseDie bool = false

func main() {
  // Set up logging. If there is a problem creating the log file, just use
  // stdout.
  f, err := os.OpenFile("log.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
  if err != nil {
    // If there was a problem, set up our logging preferences with STDOUT
    // then exit after logging the error message.
    setUpLogging(os.Stdout)
    log.Fatalf("[FATAL] Error opening log file: %s\n", err)
  }
  defer closeLogging(f)
  setUpLogging(f)

  var wg sync.WaitGroup
  quitChan := make(chan struct{})
  errChan := make(chan error)

  wg.Add(1); go connectToFAHClient(&wg, quitChan, errChan)

  // Waits for a goroutine to signal that the application should exit.
  waitForExitSignal(errChan)

  // Will signal to any unfinished goroutines that they should
  // exit cleanly.
  close(quitChan)

  // Wait for those goroutines to finish if they aren't already.
  wg.Wait()

  // Clean up.
  close(errChan)
}

func setUpLogging(f *os.File) {
  log.SetOutput(f)
  log.SetFlags(log.LstdFlags | log.Lmicroseconds)

  log.Println("========== BEGINNING LOG ==========")
}

func closeLogging(f *os.File) {
  log.Print("========== CLOSING LOG ==========\n\n")
  f.Close()
}

func waitForExitSignal(errChan <-chan error) {
  for {
    select {
    case err := <-errChan:
      if e, ok := err.(fahclient.Error); ok {
        fahclient.InspectError(e)
      } else if err.Error() == "Exiting gracefully." {
        log.Printf("[EXIT] %s\n", err)
      } else {
        log.Printf("[FATAL] Unexpected error occurred: %s\n", err)
      }
      return
    }
  }
}

func connectToFAHClient(wg *sync.WaitGroup,
                        q <-chan struct{},
                        e chan<- error) {
  defer wg.Done()

  var stage string = "connect to client"

  var conn *fahclient.Conn
  var err error

  for {
    select {
    case <-q:
      return
    default:
      switch stage {
      case "connect to client":
        conn, err = fahclient.Connect("127.0.0.1:36330", 2)
        if err != nil {
          goto returnError
        }
        defer conn.Shutdown()

        log.Println("[INFO] Connected to FAHClient.")

        stage = "read client greeting"

      case "read client greeting":
        var response []byte
        response, err = conn.ReadClient(256)
        if err != nil {
          goto returnError
        }

        if !bytes.Contains(response, []byte(fahclient.Greeting)) {
          log.Fatalln(
            "[FATAL] Don't know how to handle FAHClient response:",
            hex.EncodeToString(response),
          )
        }
        log.Println("[INFO] Received FAHClient Greeting.")

        stage = "exit"

      case "exit":
        e <- errors.New("Exiting gracefully.")
        return
      }
    }
  }
returnError:
  e <- err
  return
}
