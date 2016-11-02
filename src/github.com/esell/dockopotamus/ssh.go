package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/kr/pty"

	"golang.org/x/crypto/ssh"
)

func startSSH() {
	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() != "" && string(pass) != "" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:"+*listenPort)
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}
	log.Print("Listening on " + *listenPort + "...")

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection: ", err)
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans, sshConn.RemoteAddr().String())
	}

}

func handleChannels(chans <-chan ssh.NewChannel, remoteAddr string) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go handleChannel(newChannel, remoteAddr)
	}
}

func handleChannel(newChannel ssh.NewChannel, remoteAddr string) {
	// Since we're handling a shell, we expect a
	// channel type of "session". The also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}
	// cleanup
	remoteAddrClean := strings.Replace(remoteAddr, ":", "_", -1)
	err = os.Mkdir(*logDir+"/"+remoteAddrClean, 0777)
	if err != nil {
		log.Println("error creating log directory: ", err)
	}
	//fire up our fake shell
	bash := exec.Command("docker", "run", "-it", "--name", remoteAddrClean, "-v", *logDir+"/"+remoteAddrClean+":/var/log", *imageName, "/bin/bash")

	// Prepare teardown function
	close := func() {
		connection.Close()
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
		io.Copy(connection, bashf)
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, connection)
		once.Do(close)
	}()
	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func(in <-chan *ssh.Request) {
		for req := range in {
			switch req.Type {
			case "shell":
				// We don't accept any commands (Payload),
				// only the default shell.
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				} else {
					req.Reply(false, nil)
				}
			case "pty-req":
				// Responding 'ok' here will let the client
				// know we have a pty ready for input
				req.Reply(true, nil)
			case "window-change":
				continue //no response
			}
		}
	}(requests)
}
