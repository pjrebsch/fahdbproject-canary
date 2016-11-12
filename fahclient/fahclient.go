
package fahclient

import (
  "time"
  "net"
  "bytes"
  "log"
  "os"
  "syscall"
)

const (
  // Expected telnet greeting from the FAHClient that lets
  // us know that it's ready for communication.
  greeting string = "Welcome to the Folding@home Client command server.\n> "
)

func secDuration(secs uint8) time.Duration {
  return time.Second * time.Duration(secs)
}

type Conn struct {
  net.Conn
  Logger *log.Logger
}

func DoesContainGreeting(response []byte) bool {
  return bytes.Contains(response, []byte(greeting))
}

func Connect(addr string, tmout uint8) (*Conn, error) {
  logger := log.New(os.Stdout, addr+" -- ", log.LstdFlags | log.Lmicroseconds)

  netconn, err := net.DialTimeout("tcp", addr, secDuration(tmout))
  if err != nil {
    if netOpError, ok := err.(*net.OpError); ok {
      if osSyscallError, ok := netOpError.Err.(*os.SyscallError); ok {
        if syscallErrno, ok := osSyscallError.Err.(syscall.Errno); ok {
          switch syscallErrno {
          case syscall.ECONNREFUSED:
            logger.Fatalf("[FATAL] The connection to the client was refused!\n")
          default:
            logFatalUnknownErr("syscall.Errno", syscallErrno)
          }
        } else {
          logFatalUnknownErr("*os.SyscallError", osSyscallError)
        }
      } else {
        logFatalUnknownErr("*net.OpError", netOpError)
      }
    } else {
      logFatalUnknownErr("error", err)
    }

    return nil, err
  }

  return &Conn{netconn, logger}, nil
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
