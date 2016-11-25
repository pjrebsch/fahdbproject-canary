
package fahclient

import (
  "time"
  "net"
  "log"
)

const (
  // Expected telnet greeting from the FAHClient that lets
  // us know that it's ready for communication.
  Greeting string = "Welcome to the Folding@home Client command server.\n> "
)

func calculateWait(x int) int {
  return (x * x) + (2 * x) + 2
}

type Conn struct {
  net.Conn
}

func Connect(addr string, secs uint8) (*Conn, error) {
  timeout := time.Duration(secs) * time.Second

  var netconn net.Conn
  var err error

  for i := 1; ; i++ {
    netconn, err = net.DialTimeout("tcp", addr, timeout)
    if err != nil {
      log.Printf("[ERROR] Error on connection attempt #%d: %s\n", i, err)

      if netOpError, ok := err.(*net.OpError); ok {
        if netOpError.Temporary() {
          if i < 5 {
            wait := calculateWait(i-1)
            log.Printf("[INFO] Retrying in %d seconds...\n", wait)
            time.Sleep(time.Duration(wait) * time.Second)
            continue
          }
        }
      }

      // Don't display this if we weren't retrying.
      if i > 1 {
        log.Println("[INFO] Giving up on retries!")
      }
    }
    break
  }

  if err != nil {
    return nil, err
  }
  return &Conn{netconn}, nil
}

func (c *Conn) Shutdown() {
  c.Close()
}

func (c *Conn) ReadClient(bufSize uint16) ([]byte, error) {
  buf := make([]byte, bufSize)
  err := c.SetReadDeadline(time.Now().Add(5 * time.Second))
  // err := c.SetReadDeadline(time.Now())
  if err != nil {
    return nil, err
  }

  length, err := c.Read(buf)
  if err != nil {
    return nil, err
  }
  return buf[:length], nil
}
