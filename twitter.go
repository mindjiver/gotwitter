package main

import "fmt"
import "net/http"

func main() {
	resp, error := http.Get("http://www.twitter.com")
	if err != nil {
	}
	fmt.Println(resp)
}
