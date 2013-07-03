package main

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		// Block until a signal is received.
		s := <-c
		fmt.Println("Caught interrupt, exiting ", s)
		fmt.Println("Shutting down fake Gerrit stream-events server")
		os.Exit(1)
	}()

	hostname := "127.0.0.1"
	port := "29418"

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig {
	PasswordCallback: func(conn *ssh.ServerConn, username string, password string) bool {
			return username == "username" && password == "password"
		},
	PublicKeyCallback: func(conn *ssh.ServerConn, user, algo string, pubkey []byte) bool {
			// since we don't want to handle keys in this
			// simple server we just accept any user which
			// sends a key.
			return true
		},
	}

	pemBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		panic("Failed to load private key due to " + err.Error())
	}
	if err = config.SetRSAPrivateKey(pemBytes); err != nil {
		panic("Failed to parse private key")
	}

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := ssh.Listen("tcp", string(hostname+":"+port), config)
	if err != nil {
		panic("failed to listen for connection")
	}

	fmt.Fprintf(os.Stdout, "Started fake Gerrit stream-event server on %s port %s\n", hostname, port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic("failed to accept incoming connections")
		}

		if err := conn.Handshake(); err != nil {
			fmt.Println("Failed to handshake with client")
		}

		// A ServerConn multiplexes several channels, which must
		// themselves be Accepted.

		for {
			// Accept reads from the connection, demultiplexes packets
			// to their corresponding channels and returns when a new
			// channel request is seen. Some goroutine must always be
			// calling Accept; otherwise no messages will be forwarded
			// to the channels.
			channel, err := conn.Accept()
			if err != nil || channel == nil {
				fmt.Println("Failed to accept connection")
				fmt.Println("This error: ", err)
				break
			}

			// Channels have a type, depending on the application level
			// protocol intended. In the case of a shell, the type is
			// "session" and ServerShell may be used to present a simple
			// terminal interface.
			if channel.ChannelType() != "session" {
				channel.Reject(ssh.UnknownChannelType, "unknown channel type")
				continue
			}
			channel.Accept()

			term := terminal.NewTerminal(channel, "")
			serverTerm := &ssh.ServerTerminal{
				Term:    term,
				Channel: channel,
			}

			go func() {
				defer channel.Close()
				for {
					command, err := serverTerm.ReadLine()
					if err != nil {
						break
					}
					example_json_stream := "{\"type\":\"comment-added\",change:{\"project\":\"tools/gerrit\", ...}, ...}\n"
					switch {
					case command == "gerrit stream-events":

						serverTerm.Write([]byte(example_json_stream))
					default:
						serverTerm.Write([]byte(command))
					}

				}
			}()
		}
	}
}
