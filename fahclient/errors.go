
package fahclient

import (
  "net"
  "syscall"
  "os"
  "fmt"
)

func describeUnknownErr(errType string, err error) error {
  return fmt.Errorf(
    "Don't know how to handle %v (%T): %v",
    errType,
    err,
    err,
  )
}

func TranslateConnectionError(err error) error {
  var newErr error

  if netOpError, ok := err.(*net.OpError); ok {
    if osSyscallError, ok := netOpError.Err.(*os.SyscallError); ok {
      if syscallErrno, ok := osSyscallError.Err.(syscall.Errno); ok {
        switch syscallErrno {
        case syscall.ECONNREFUSED:
          newErr = fmt.Errorf(
            "The connection to the FAHClient was refused. "+
            "Please ensure that the FAHClient is running on the "+
            "host and port %s and that the port is not being blocked "+
            "by a firewall.",
            netOpError.Addr.String(),
          )
        default:
          newErr = describeUnknownErr("syscall.Errno", syscallErrno)
        }
      } else {
        newErr = describeUnknownErr("*os.SyscallError", osSyscallError)
      }
    } else {
      newErr = describeUnknownErr("*net.OpError", netOpError)
    }
  } else {
    newErr = describeUnknownErr("error", err)
  }

  return newErr
}
