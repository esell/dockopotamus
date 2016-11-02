package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

var (
	imageName  = flag.String("i", "esell/dockopotamus", "Docker image to use (must be pulled already)")
	listenPort = flag.String("p", "22", "Port to listen on")
	keyFile    = flag.String("k", "id_rsa", "Private key file to use for server")
	logDir     = flag.String("l", "/logs", "Log directory (will be created if it doesn't exist)")
)

func main() {
	flag.Parse()
	go startSSH()
	go startTelnet()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			return
		}
	}
}
