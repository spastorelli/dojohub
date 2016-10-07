package components

import (
	"bytes"
	"net/http"
	"os"
	"strconv"
)

// staticFileSystem implements a http.FileSystem that only serve static files.
type staticFileSystem struct {
	fs http.FileSystem
}

func (sfs staticFileSystem) Open(name string) (http.File, error) {
	f, err := sfs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	d, err := f.Stat()
	if d.IsDir() {
		return nil, os.ErrNotExist
	}
	return f, nil
}

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
	staticDir    string
	certFile     string
	keyFile      string
}

// NewAppServer returns a new App Server setup to listen on the given hostname and port
// to handle HTTP connections. If the certFile and keyFile are provided the server is
// setup to handle HTTPS connections.
func NewAppServer(hostname *string, port *int, staticDir, certFile, keyFile *string) *AppServer {
	hostAddr := buildServerAddress(*hostname, *port)
	return &AppServer{
		&http.Server{
			Addr:    hostAddr,
			Handler: http.DefaultServeMux,
		}, *staticDir, *certFile, *keyFile}
}

// useTLS checks if the App Server is setup to handle HTTPS connections.
func (s *AppServer) useTLS() bool {
	return s.certFile != "" && s.keyFile != ""
}

func (s *AppServer) serveStatic() (err error) {
	fs := http.FileServer(staticFileSystem{http.Dir(s.staticDir)})
	if sMux, ok := s.Handler.(*http.ServeMux); ok {
		sMux.Handle("/static/", http.StripPrefix("/static/", fs))
	}
	return
}

// RegisterHandler registers a HTTP handler for a given pattern.
func (s *AppServer) RegisterHandler(pattern string, handler http.Handler) {
	if sMux, ok := s.Handler.(*http.ServeMux); ok {
		sMux.Handle(pattern, handler)
	}
}

// Serve starts the App Server to listen for incoming connections.
func (s *AppServer) Serve() (err error) {
	s.serveStatic()

	if s.useTLS() {
		err = s.ListenAndServeTLS(s.certFile, s.keyFile)
	} else {
		err = s.ListenAndServe()
	}
	return
}
