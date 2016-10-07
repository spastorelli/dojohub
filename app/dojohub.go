package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/spastorelli/dojohub/app/components"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Global variables to hold flags values.
var (
	host                           string
	port                           int
	staticDir                      string
	tlsCertFile, tlsPrivateKeyFile string
)

// TODO(spastorelli): Remove when actual handlers are implemented.
func fakeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello There!")
}

// handleTermination handles gracefully the termination via SIGINT, SIGTERM of the server.
func handleTermination() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signals
		glog.Infof("Signal %v intercepted. Exiting...", sig)
		glog.Flush()
		os.Exit(1)
	}()
}

func init() {
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.StringVar(&host, "host", "127.0.0.1", "HTTP server host")
	flag.StringVar(&staticDir, "staticDir", "static/", "The static files directories")
	flag.StringVar(
		&tlsCertFile, "tlsCertFile", "",
		"The file that contains the TLS/SSL certificate for the server.")
	flag.StringVar(
		&tlsPrivateKeyFile, "tlsPrivateKeyFile", "",
		"The file that contains the TLS/SSL private key for the server.")
}

func main() {
	flag.Parse()
	handleTermination()

	app := components.NewAppServer(&host, &port, &staticDir, &tlsCertFile, &tlsPrivateKeyFile)
	app.RegisterHandler("/", http.HandlerFunc(fakeHandler))

	if err := app.Serve(); err != nil {
		glog.Error("Error while starting the DojoHub: ", err)
	}
}
