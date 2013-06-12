package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func http_request(url string) []byte {
	response, error := http.Get(url)
	if error != nil {
		fmt.Println("Could not connect to server and retrieve data, exiting.\n%s",
			error)
		os.Exit(1)
	}

	defer response.Body.Close()
	contents, error := ioutil.ReadAll(response.Body)
	if error != nil {
		fmt.Println("Failed to read HTTP response from server, exiting.\n%s",
			error)
		os.Exit(1)
	}

	return contents

}

func main() {
	url := "https://twitter.com"
        contents := http_request(url)
	fmt.Println("%s", string(contents))
}
