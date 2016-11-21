
package fahclient

import (
  "time"
  "net"
  "bytes"
  "log"
  "os"
)

const (
  // Expected telnet greeting from the FAHClient that lets
  // us know that it's ready for communication.
  greeting string = "Welcome to the Folding@home Client command server.\n> "
)

func calculateWait(x int) int {
  return (x * x) + (2 * x) + 2
}

type Conn struct {
  net.Conn
  Logger *log.Logger
}

func DoesContainGreeting(response []byte) bool {
  return bytes.Contains(response, []byte(greeting))
}

func Connect(addr string, secs uint8) *Conn {
  timeout := time.Duration(secs) * time.Second
  logger := log.New(os.Stdout, addr+" ", log.LstdFlags | log.Lmicroseconds)

  var netconn net.Conn
  var err error

  for i := 1; ; i++ {
    netconn, err = net.DialTimeout("tcp", addr, timeout)
    if err != nil {
      logger.Printf("[ERROR] Error on connection attempt #%d: %s\n", i, err)

      if netOpError, ok := err.(*net.OpError); ok {
        if netOpError.Temporary() {
          if i < 5 {
            wait := calculateWait(i-1)
            logger.Printf("[INFO] Retrying in %d seconds...\n", wait)
            time.Sleep(time.Duration(wait) * time.Second)
            continue
          }
        }
      }

      // Don't display this if retrying wasn't going to happen anyway.
      if i > 1 {
        logger.Println("[INFO] Giving up on retries!")
      }
    }
    break
  }

  if err != nil {
    fatalInspectError(err, logger)
    return nil
  }

  return &Conn{netconn, logger}
}

func (c *Conn) Shutdown() {
  c.Close()
}

func (c *Conn) ReadClient(bufSize uint16) ([]byte, error) {
  buf := make([]byte, bufSize)
  length, err := c.Read(buf)
  if err != nil {
    return nil, nil
  }
  return buf[:length], nil
}
