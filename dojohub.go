package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/spastorelli/dojohub/components"
	"github.com/spastorelli/dojohub/handlers"
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
	flag.StringVar(&host, "host", "", "HTTP server host")
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

	chatExampleApp := components.NewApplication(
		os.Getenv("APP_ID"),
		"ChatExampleApp",
		os.Getenv("APP_SECRET"),
	)
	msgHub := components.NewMsgHub()
	msgHub.RegisterApplication(chatExampleApp)
	msgHub.Run()

	app := components.NewAppServer(&host, &port, &staticDir, &tlsCertFile, &tlsPrivateKeyFile)
	app.RegisterHandler("/", http.HandlerFunc(handlers.Home))
	app.RegisterHandler("/example/chat/", http.HandlerFunc(handlers.ExampleChat))
	app.RegisterHandler("/ws/", msgHub)

	if err := app.Serve(); err != nil {
		glog.Error("Error while starting the DojoHub: ", err)
	}
}
