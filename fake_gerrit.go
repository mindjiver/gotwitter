package main

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"log"
	"strings"
)

var logFile *os.File

func main() {
	example_json_stream := "{\"type\":\"patchset-created\",\"change\":{\"project\":\"example\",\"branch\":\"master\",\"id\":\"I44f1ae9dbc886cddfa108b47849e2e1b83b548cd\",\"number\":\"7\",\"subject\":\"Remove newline\",\"owner\":{\"name\":\"Peter Jönsson\",\"email\":\"peter.jonsson@klarna.com\",\"username\":\"peter.jonsson\"},\"url\":\"http://localhost:8082/r/7\",\"status\":\"NEW\"},\"patchSet\":{\"number\":\"1\",\"revision\":\"44f1ae9dbc886cddfa108b47849e2e1b83b548cd\",\"parents\":[\"0009ab1c17c24d8bfcfad0b67a06c424cc02e487\"],\"ref\":\"refs/changes/07/7/1\",\"uploader\":{\"name\":\"Peter Jönsson\",\"email\":\"peter.jonsson@klarna.com\",\"username\":\"peter.jonsson\"},\"createdOn\":1372846590,\"author\":{\"name\":\"Peter Jönsson\",\"email\":\"peter.joensson@gmail.com\",\"username\":\"\"},\"sizeInsertions\":0,\"sizeDeletions\":-1},\"uploader\":{\"name\":\"Peter Jönsson\",\"email\":\"peter.jonsson@klarna.com\",\"username\":\"peter.jonsson\"}}"

	var err error
	logFile, err = os.Create("fakegerrit.log")
	if err != nil {
		log.Fatal("Log file create:", err.Error())
		return
	}
	defer logFile.Close()

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
		iteration := 0
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
			fmt.Println("Successfully accepted channel from client")

			term := terminal.NewTerminal(channel, "")
			serverTerm := &ssh.ServerTerminal {
				Term:    term,
				Channel: channel,
			}

			go func() {
				defer channel.Close()
				for {
					iteration = iteration + 1
					fmt.Println("Reading from SSH console", iteration)
					command, err := serverTerm.ReadLine()
					if err != nil {
						break
					}
					fmt.Fprintf(os.Stdout, "received: %s\n", command)
					fmt.Fprintf(logFile,   "received: %s\n", command)
					switch {
					case command == "gerrit stream-events":
						serverTerm.Write([]byte(example_json_stream + "\n"))
					case command == "gerrit version":
						serverTerm.Write([]byte("gerrit version 2.6"))
					case strings.Contains(command, "query"):
						status_open := "{\"project\":\"example\",\"branch\":\"master\",\"id\":\"I960171c79c0456d470b6e80114c39e1ea6fd615d\",\"number\":\"8\",\"subject\":\"Fix dumpy dooo\",\"owner\":{\"name\":\"Peter Jönsson\",\"email\":\"peter.jonsson@klarna.com\",\"username\":\"peter.jonsson\"},\"url\":\"http://localhost:8082/r/8\",\"createdOn\":1372853084,\"lastUpdated\":1372853084,\"sortKey\":\"002627d400000008\",\"open\":true,\"status\":\"NEW\"}"
						serverTerm.Write([]byte(status_open))
						serverTerm.Write([]byte("{\"type\":\"stats\",\"rowCount\":1,\"runTimeMilliseconds\":10}"))
					default:
						break
					}

				}
			}()
		}
	}
}
