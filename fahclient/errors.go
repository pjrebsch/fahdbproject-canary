
package fahclient

import (
  "log"
  "net"
  "syscall"
  "os"
)

func logFatalUnknownErr(errType string, err error) {
  log.Fatalf(
    "[FATAL] Not sure how to handle %v (%T): %v\n\n",
    errType,
    err,
    err,
  )
}

func fatalInspectError(err error, logger *log.Logger) {
  if netOpError, ok := err.(*net.OpError); ok {
    if osSyscallError, ok := netOpError.Err.(*os.SyscallError); ok {
      if syscallErrno, ok := osSyscallError.Err.(syscall.Errno); ok {
        switch syscallErrno {
        case syscall.ECONNREFUSED:
          logger.Fatalf(
            "[FATAL] The connection to the FAHClient was refused. "+
            "Please ensure that the FAHClient is running on the "+
            "host and port %s and that the port is not being blocked "+
            "by a firewall.\n\n",
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
