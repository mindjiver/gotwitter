package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	service := ":9000"
	tcpAddr, error := net.ResolveTCPAddr("ipv4", service)
	check_error(error)

	listener, error := net.ListenTCP("tcp", tcpAddr)
	check_error(error)

	for {
		conn, error := listener.Accept()
		if error != nil {
			continue
		}

		go handle_connection(conn)
	}
}

func handle_connection(conn net.Conn) {
	// close connection on exit
	defer conn.Close()
	buffer := make([]byte, 1024)
	length, error := conn.Read(buffer)
	if error != nil {
		// do something good to clean up?
	} else {
		// strip away newline and null termination
		command := string(buffer[:length-2])
		switch {

		case command == "gerrit stream-events":
			conn.Write([]byte("gerrit subcommand\n"))
		default:
			conn.Write([]byte(command))
		}
	}
}

func check_error(error error) {
	if error != nil {
		fmt.Fprintf(os.Stderr, "Fatal error %s", error.Error())
		os.Exit(1)
	}
}
