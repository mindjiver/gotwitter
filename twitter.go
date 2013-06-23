package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

func http_request(url string) []byte {
	response, error := http.Get(url)
	check_error(error)

	defer response.Body.Close()
	contents, error := ioutil.ReadAll(response.Body)
	check_error(error)

	return contents
}

func main() {
	// 29418 is the default Gerrit Port, here we should send git
	// events when receving the "stream-events" command.


	//url := "https://twitter.com"
        //contents := http_request(url)
	//fmt.Println("%s", string(contents))

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
	_, error := conn.Read(buffer)
	if error != nil {
		// do something good to clean up?
	} else {
		switch {
		case string(buffer) == "gerrit":
			conn.Write([]byte("gerrit command"))
		default:
			conn.Write(buffer[0:])
		}
	}
}

func check_error(error error) {
	if error != nil {
		fmt.Fprintf(os.Stderr, "Fatal error %s", error.Error())
		os.Exit(1)
	}
}
