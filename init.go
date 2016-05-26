// grace provides for graceful restart for go http servers.
// There are 2 parts to graceful restarts
// 1. Share listening sockets (this is done via socketmaster binary)
// 2. Close listener gracefully (via graceful)
package grace

import (
	graceful "gopkg.in/tylerb/graceful.v1"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
  "flag"
	"time"
)

var listenPort string

// add -p flag to the list of flags supported by the app,
// and allow it to over-ride default listener port in config/app
func init() {
  flag.StringVar(&listenPort,"p","","listener port")
}

// applications need some way to access the port
// TODO: this method will work only after grace.Serve is called.
func GetListenPort(hport string) string {
  return listenPort;
}

// start serving on hport. If running via socketmaster, the hport argument is
// ignored. Also, if a port was specified via -p, it takes precedence on hport
func Serve(hport string, handler http.Handler) error {
  l,err := Listen(hport)
  if err != nil {
		log.Fatalln(err)
  }

	srv := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Handler: handler,
		},
	}

	log.Println("starting serve on ", hport)
	return srv.Serve(l)
}

// This method can be used for any TCP Listener, e.g. non HTTP
func Listen(hport string) (net.Listener,error) {
	var l net.Listener

	fd := os.Getenv("EINHORN_FDS")
	if fd != "" {
		sock, err := strconv.Atoi(fd)
		if err == nil {
      hport = "socketmaster:" + fd
			log.Println("detected socketmaster, listening on", fd)
			file := os.NewFile(uintptr(sock), "listener")
			fl, err := net.FileListener(file)
			if err == nil {
				l = fl
			}
		}
	}

  if listenPort != "" {
    hport = ":" + listenPort
  }

	if l == nil {
		var err error
		l, err = net.Listen("tcp4", hport)
		if err != nil {
      return nil,err
		}
	}

  return l,nil
}
