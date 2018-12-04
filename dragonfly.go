package main

import (
	"flag"
	"fmt"
	cfg "github.com/NBCFB/Dragonfly/config"
	ls "github.com/NBCFB/Dragonfly/listener"
	"github.com/husobee/vestigo"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	//var wait time.Duration
	//flag.DurationVar(&wait, "graceful-timeout", time.Second*60,
	//	"the duration for which the server gracefully wait for existing connections to finish - 1 minute.")
	//flag.Parse()

	r := vestigo.NewRouter()
	vestigo.AllowTrace = true

	// Setting up router global CORS policy
	r.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:		[]string{"*"},
		AllowCredentials:	true,
		AllowMethods:		[]string{"GET", "POST"},
		AllowHeaders:		[]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token",
			"Authorization", "X-Requested-With"},
		ExposeHeaders:		[]string{"Accept", "Content-Type", "Content-Length"},
		MaxAge:				3600 * time.Second,
	})


	// Read configuration
	err := cfg.Reader()
	if err != nil {
		log.Printf("Unable to read _config file! %v", err.Error())
	}

	var address string
	mode := viper.GetString("Mode")
	host := viper.GetString(fmt.Sprintf("%v.%v", mode, "host"))
	flag.StringVar(&address, "addr", host + ":8787", "Address to listen on.")

	// Create (or import) a net.Listener and start a goroutine that runs
	// a HTTP server on that net.Listener.
	listener, err := ls.CreateOrImportListener(address)
	if err != nil {
		log.Printf("Unable to create or import a http _listener: %v.\n", err)
		os.Exit(1)
	}
	server := startServer(address, listener, r)

	// Wait for signals to either fork or quit.
	err = ls.WaitForSignals(address, listener, server)
	if err != nil {
		log.Printf("Exiting NBCFB-Dragonfly: %v\n", err)
		return
	}
	log.Printf("Exiting NBCFB-Dragonfly [PID:%v].\n", os.Getpid())
}

func startServer(addr string, ln net.Listener, router *vestigo.Router) *http.Server {

	httpServer := &http.Server{
		//Addr: "0.0.0.0:8787",
		Addr: addr,
		//Set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	go httpServer.Serve(ln)
	log.Printf("HTTP Server NBCFB-Dragonfly started [PID:%v].\n", os.Getpid())

	return httpServer
}