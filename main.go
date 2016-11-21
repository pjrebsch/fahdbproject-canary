
package main

import (
  "github.com/pjrebsch/fahdbproject-canary/fahclient"
  "log"
)

func main() {
  conn := fahclient.Connect("127.0.0.1:36331", 2)
  defer conn.Shutdown()

  response, err := conn.ReadClient(256)
  if err != nil {
    panic(err)
  }

  if !fahclient.DoesContainGreeting(response) {
    log.Fatalln("Unknown how to handle FAHClient response:", response)
  }
  log.Println("Received FAHClient Greeting.")
}
