package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andelf/go-curl"
	"log"
)

const (
	endpoint = "https://sandbox.api.realogy.com"
)
// this is a simple curl client using golang for demonstrating streaming and changes api clients.
func main() {
	c := curl.EasyInit()
	defer c.Cleanup()

	// 1. change stream api client with a timeout of 10 seconds
	callApi(c, endpoint+"/mls/changes", []string{"Grpc-Timeout: 10S"})

	// 2. stream listings by source api with the timeout of 5 seconds
	callApi(c, endpoint+"/mls/stream/source/AR_TMLS", []string{"Grpc-Timeout: 5S"})

	// 3. change stream api client with "apiKey" and "Authorization" headers.
	// callApi(c, endpoint+"/mls/changes", []string{"Grpc-Timeout: 10S", "apiKey: abc123", "Authorization: Bearer abc123"})

	// 4. change stream api client that handles error after a configured timeout ("Grpc-Timeout") and configured number of attempts (count)
	count := 5 // number of attempts to listen for changes.
	for ;count > 0; {
		count--
		log.Print("Calling mls api to receive listing changes")
		changesErr := callApi2(c, endpoint+"/mls/changes", []string{"Grpc-Timeout: 5S"}) // 5secs timeout
		var open bool
		var errData error
		if errData, open = <-changesErr; open {
			if errData != nil {
				var response map[string]interface{}
				if err := json.Unmarshal([]byte(errData.Error()), &response); err != nil {
					panic(err)
				}
				if response["error"] != "" && response["code"] == float64(4) {
					log.Printf("Request timeout with error: %v", response)
				}
			}
		}
	}
}

// generic streaming api that accepts endpoint and slice of headers as parameters.
func callApi(c *curl.CURL, endpoint string, headers []string) {
	c.Setopt(curl.OPT_URL, endpoint)
	c.Setopt(curl.OPT_HTTPHEADER, headers)

	changeStream := func (buf []byte, data interface{}) bool {
		println("Response: ", string(buf))
		return true
	}
	c.Setopt(curl.OPT_WRITEFUNCTION, changeStream)

	if err := c.Perform(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// generic streaming api that accepts endpoint and slice of headers as parameters. returns error channel.
func callApi2(c *curl.CURL, endpoint string, headers []string) <-chan error {
	c.Setopt(curl.OPT_URL, endpoint)
	c.Setopt(curl.OPT_HTTPHEADER, headers)

	changeStream := func (buf []byte, data interface{}) bool {
		ch, ok := data.(chan string)
		if ok {
			ch <- string(buf)
			return true
		} else {
			println("Error sending the data!")
			return false
		}
		return false
	}
	c.Setopt(curl.OPT_WRITEFUNCTION, changeStream)

	// make a chan
	ch := make(chan string, 1)
	errs := make(chan error, 1)
	go func(ch chan string) {
		for {
			data := <-ch
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(data), &response); err != nil {
				panic(err)
			}
			if response["error"] != "" {
				errs <- errors.New(data)
			}
		}
	}(ch)

	c.Setopt(curl.OPT_WRITEDATA, ch)

	if err := c.Perform(); err != nil {
		fmt.Printf("Error during curl: %v\n", err)
	}

	return errs
}
