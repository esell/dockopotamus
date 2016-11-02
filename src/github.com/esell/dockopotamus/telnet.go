package main

import (
	"io"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"

	"github.com/kr/pty"
)

func startTelnet() {
	listener, err := net.Listen("tcp", "0.0.0.0:23")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}
	log.Print("Listening on 23...")

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection: ", err)
		}
		go handleChannelTelnet(nConn)
	}
}

func handleChannelTelnet(c net.Conn) {
	// cleanup
	remoteAddrClean := strings.Replace(c.RemoteAddr().String(), ":", "_", -1)
	//fire up our fake shell
	bash := exec.Command("docker", "run", "-it", "--name", remoteAddrClean+"_telnet", "-v", *logDir+"/"+remoteAddrClean+"_telnet:/var/log", *imageName, "/bin/bash")

	// Prepare teardown function
	close := func() {
		c.Close()
		_, err := bash.Process.Wait()
		if err != nil {
			log.Printf("Failed to exit bash (%s)", err)
		}
		log.Printf("Session closed")
	}

	// Allocate a terminal for this channel
	log.Print("Creating pty...")
	bashf, err := pty.Start(bash)
	if err != nil {
		log.Printf("Could not start pty (%s)", err)
		close()
		return
	}

	//pipe session to bash and visa-versa
	var once sync.Once
	go func() {
		io.Copy(c, bashf)
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, c)
		once.Do(close)
	}()
}
