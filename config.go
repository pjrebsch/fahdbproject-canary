
package main

import (
  "fmt"
  "encoding/json"
  "io"
  "io/ioutil"
  "os"
  "bytes"
)

func loadConfig() error {
  f, err := os.Open("config.json")
  if err != nil {
    return err
  }
  defer f.Close()

  // Prevent an attack where the config file size has been made
  // unreasonably large and this application consumes that amount
  // of memory. 64KiB should be more than enough and still not
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
