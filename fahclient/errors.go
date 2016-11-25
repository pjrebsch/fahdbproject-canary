
package fahclient

import (
  "log"
  "net"
  "syscall"
  "os"
)

type Error struct {
  Err error
}

func (e Error) Error() string {
  return e.Err.Error()
}

func logFatalUnknownErr(errType string, err error) {
  log.Fatalf(
    "[FATAL] Don't know how to handle %v (%T): %v\n",
    errType,
    err,
    err,
  )
}

func InspectError(origErr Error) {
  // Extract the real error from the custom error.
  err, _ := origErr.Err.(error)

  if netOpError, ok := err.(*net.OpError); ok {
    if osSyscallError, ok := netOpError.Err.(*os.SyscallError); ok {
      if syscallErrno, ok := osSyscallError.Err.(syscall.Errno); ok {
        switch syscallErrno {
        case syscall.ECONNREFUSED:
          log.Printf(
            "[FATAL] The connection to the FAHClient was refused. "+
            "Please ensure that the FAHClient is running on the "+
            "host and port %s and that the port is not being blocked "+
            "by a firewall.\n",
            netOpError.Addr.String(),
          )
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
}
