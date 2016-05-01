package main

import (
	"bytes"
	"net/http"
	"strconv"
)

// buildServerAddress returns the host address based on a hostname and port.
func buildServerAddress(hostname string, port int) string {
	var addr bytes.Buffer
	addr.WriteString(hostname)
	addr.WriteString(":")
	addr.WriteString(strconv.Itoa(port))

	return addr.String()
}

// AppServer provides basic abstraction to setting up a *http.Server.
type AppServer struct {
	*http.Server // anonymous underlying http.Server
	certFile     string
	keyFile      string
}

// NewAppServer returns a new App Server setup to listen on the given hostname and port
// to handle HTTP connections. If the certFile and keyFile are provided the server is
// setup to handle HTTPS connections.
func NewAppServer(hostname *string, port *int, certFile, keyFile *string) *AppServer {
	hostAddr := buildServerAddress(*hostname, *port)
	return &AppServer{
		&http.Server{
			Addr:    hostAddr,
			Handler: http.DefaultServeMux,
		}, *certFile, *keyFile}
}

// useTLS checks if the App Server is setup to handle HTTPS connections.
func (s *AppServer) useTLS() bool {
	return s.certFile != "" && s.keyFile != ""
}

// RegisterHandler registers a HTTP handler for a given pattern.
func (s *AppServer) RegisterHandler(pattern string, handler http.Handler) {
	if sMux, ok := s.Handler.(*http.ServeMux); ok {
		sMux.Handle(pattern, handler)
	}
}

// Serve starts the App Server to listen for incoming connections.
func (s *AppServer) Serve() (err error) {
	if s.useTLS() {
		err = s.ListenAndServeTLS(s.certFile, s.keyFile)
	} else {
		err = s.ListenAndServe()
	}
	return
}
