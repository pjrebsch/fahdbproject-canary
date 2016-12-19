
package main

import (
  "log"
  "os"
)

func setUpLogging(f *os.File) {
  log.SetOutput(f)
  log.SetFlags(log.LstdFlags | log.Lmicroseconds)

  log.Println("========== BEGINNING LOG ==========")
}

func closeLogging(f *os.File) {
  log.Print("=========== CLOSING LOG ===========\n\n")
  f.Close()
}
