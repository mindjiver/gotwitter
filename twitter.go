package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	url := "https://twitter.com"
	response, error := http.Get(url)
	if error != nil {
		fmt.Println("Could not connect to server and retrieve data, exiting")
		fmt.Println(error)
		os.Exit(1)
	}

	defer response.Body.Close()
	contents, error := ioutil.ReadAll(response.Body)
	if error != nil {
		fmt.Println("Failed to read HTTP response from server, exiting")
		fmt.Println(error)
		os.Exit(1)
	}

	fmt.Println("%s", string(contents))
}
