
package main

import (
  "github.com/pjrebsch/fahdbproject-canary/fahclient"
  "os"
  "log"
  "sync"
  "bytes"
  "encoding/hex"
  "fmt"
)

// "Flag" struct for signaling that the application should exit
// due to either the user's request or an unrecoverable error.
//
// We opt to use a global variable and mutex here because it
// is simply easier to implement than trying to work with a
// channel.
var death struct {
  sync.Mutex
  effective bool
}

type AppConfig struct {
  FAHClientHostAndPort string
}
var DefaultConfig = AppConfig{
  FAHClientHostAndPort: "127.0.0.1:36330",
}
var Config AppConfig

func main() {
  // Load config file. This should happen first so that the user
  // can customize their log output file location.
  for {
    err := loadConfig()
    if err != nil {
      if os.IsNotExist(err) {
        log.Println(
          "[INFO] No config file found. " +
          "Creating one with default config values...",
        )
        err = writeConfig(DefaultConfig)
        if err != nil {
          log.Fatalf("[FATAL] Error writing config file: %s\n\n", err)
        }
        continue
      } else {
        log.Fatalf("[FATAL] Error reading config file: %s\n\n", err)
      }
    }
    break
  }

  // Create the log file.
  f, err := os.OpenFile(
    "log.txt",
    os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
    0644,
  )
  if err != nil {
    // If there was a problem, set up logging with STDOUT
    // then exit after logging the error message.
    setUpLogging(os.Stdout)
    log.Fatalf("[FATAL] Error opening log file: %s\n\n", err)
  }
  setUpLogging(f)
  defer closeLogging(f)

  var wg sync.WaitGroup
  errors := make(chan error, 10)

  wg.Add(1); go connectToFAHClient(&wg, errors)

  // Wait for those goroutines to finish.
  wg.Wait()

  // Now that there are no more goroutines sending errors, close
  // the channel (otherwise the range will block below).
  close(errors)

  // Log all unrecoverable errors returned by the goroutines.
  for err = range errors {
    log.Printf("[FATAL] %s\n", err)
  }
}

func connectToFAHClient(wg *sync.WaitGroup, errors chan<- error) {
  defer wg.Done()
  defer signalDeath()

  conn, err := fahclient.Connect(Config.FAHClientHostAndPort, 2)
  if err != nil {
    errors <- err
    return
  }
  defer conn.Shutdown()
  log.Println("[INFO] Connected to FAHClient.")

  if death.effective { return }

  response, err := conn.ReadClient(256)
  if err != nil {
    errors <- err
    return
  }

  if !bytes.Contains(response, []byte(fahclient.Greeting)) {
    errors <- fmt.Errorf(
      "Don't know how to handle FAHClient response: ",
      hex.EncodeToString(response),
    )
    return
  }
  log.Println("[INFO] Received FAHClient Greeting.")

  return
}

func signalDeath() {
  death.Lock()
  death.effective = true
  death.Unlock()
}
