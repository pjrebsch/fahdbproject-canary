
package main

import (
  "github.com/pjrebsch/fahdbproject-canary/fahclient"
  "os"
  "log"
  "sync"
  "bytes"
  "encoding/hex"
  "encoding/json"
  "fmt"
  "io"
  "io/ioutil"
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

  f, err := os.OpenFile(
    "log.txt",
    os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
    0644,
  )
  if err != nil {
    // If there was a problem, set up our logging preferences with STDOUT
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

func setUpLogging(f *os.File) {
  log.SetOutput(f)
  log.SetFlags(log.LstdFlags | log.Lmicroseconds)

  log.Println("========== BEGINNING LOG ==========")
}

func closeLogging(f *os.File) {
  log.Print("=========== CLOSING LOG ===========\n\n")
  f.Close()
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

func loadConfig() error {
  f, err := os.Open("config.json")
  if err != nil {
    return err
  }
  defer f.Close()

  // Prevent an attack where the config file size has been made
  // unreasonably large and this application consumes that amount
  // of memory. 64KB should be more than enough and still not
  // cause any issues.
  lr := io.LimitedReader{f, 65535}

  chunkSize := 255
  var confJSON []byte

  // Read from the file until we can unmarshal the JSON successfully.
  //
  // We don't need to worry about how much we've read because
  // io.LimitedReader.Read() will return an io.EOF error if
  // its max length has been read.
  for {
    buf := make([]byte, chunkSize)
    length, readErr := lr.Read(buf)
    // If the read error is an io.EOF, we'll make sure later to
    // return an error if we still can't unmarshal the JSON
    // we've already retrieved.
    if readErr != nil && readErr != io.EOF {
      return readErr
    }

    confJSON = append(confJSON, buf[:length]...)

    jsonErr := json.Unmarshal(confJSON, &Config)
    if jsonErr != nil {
      switch jsonErr.(type) {
      case *json.SyntaxError:
        if readErr == io.EOF {
          if lr.N == 0 {
            return fmt.Errorf(
              "The config file has exceeded the maximum allowed size.",
            )
          } else {
            e := jsonErr.(*json.SyntaxError)
            return fmt.Errorf(
              "The config file is invalid JSON: %v @ offset: %v",
              e.Error(), e.Offset,
            )
          }
        }
        // We must still not have enough of the file for it to be
        // valid JSON.
        continue
      default:
        return jsonErr
      }
    }

    return nil
  }
}

func writeConfig(conf AppConfig) error {
  confJSON, err := json.Marshal(conf)
  if err != nil {
    return err
  }

  var bufConfJSON bytes.Buffer
  err = json.Indent(&bufConfJSON, confJSON, "", "\t")
  if err != nil {
    return err
  }
  indentedConfJSON := make([]byte, bufConfJSON.Len())
  _, err = bufConfJSON.Read(indentedConfJSON)
  if err != nil {
    return err
  }

  err = ioutil.WriteFile("config.json", indentedConfJSON, 0644)
  if err != nil {
    return err
  }
  return nil
}
