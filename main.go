
package main

import (
  "github.com/pjrebsch/fahdbproject-canary/fahclient"
  "os"
  "log"
  "sync"
)

func main() {
  // Set up logging. If there is a problem creating the log file, just use
  // stdout.
  f, err := os.OpenFile("log.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
  if err != nil {
    // If there was a problem, set up our logging preferences with STDOUT
    // then exit after logging the error message.
    setUpLogging(os.Stdout)
    log.Fatalf("[FATAL] Error opening log file: %s\n\n", err)
  }
  defer closeLogging(f)
  setUpLogging(f)

  var wg sync.WaitGroup
  quitChan := make(chan struct{})
  errChan := make(chan error)

  wg.Add(1); go connectFAHClient(&wg, quitChan, errChan)

  wg.Wait()
}

func setUpLogging(f *os.File) {
  log.SetOutput(f)
  log.SetFlags(log.LstdFlags | log.Lmicroseconds)

  log.Println("========== BEGINNING LOG ==========")
}

func closeLogging(f *os.File) {
  log.Println("========== CLOSING LOG ==========")
  f.Close()
}

func connectFAHClient(wg *sync.WaitGroup, q chan struct{}, e chan error) {
  defer wg.Done()

  conn, err := fahclient.Connect("127.0.0.1:36330", 2)
  if err != nil {
    fahclient.InspectError(err)
    return
  }
  defer conn.Shutdown()

  conn.ReadForGreeting()
}
