package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/armon/go-socks5"
)

// DNSAnswer stores a member of answers[] in the StatDNS result.
type DNSAnswer struct {
	Name     string
	Type     string
	Class    string
	TTL      int
	RDLength int
	RData    string
}

// DNSQuestion stores a member of question[] in the StatDNS result.
type DNSQuestion struct {
	Name  string
	Type  string
	Class string
}

// DNSResult is our type which matches the JSON object.
type DNSResult struct {
	Question []DNSQuestion
	Answer   []DNSAnswer
}

// This function fetch the content of a URL will return it as an
// array of bytes if retrieved successfully.
func getContent(url string) ([]byte, error) {
	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Send the request via a client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// Defer the closing of the body
	defer resp.Body.Close()
	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// At this point we're done - simply return the bytes
	return body, nil
}

// RunDNSLookup will attempt to get the IP record for
// a given IP. If no errors occur, it will return a pair
// of the record and nil. If it was not successful, it will
// return a pair of nil and the error.
func RunDNSLookup(name string) (*DNSResult, error) {
	// Fetch the JSON content for that given IP
	content, err := getContent(
		fmt.Sprintf("http://api.statdns.com/%s/a", name))
	if err != nil {
		// An error occurred while fetching the JSON
		return nil, err
	}
	// Fill the record with the data from the JSON
	var record DNSResult
	err = json.Unmarshal(content, &record)
	if err != nil {
		// An error occurred while converting our JSON to an object
		return nil, err
	}
	return &record, err
}

// StatDNSResolver uses the system DNS to resolve host names
type StatDNSResolver struct{}

// Resolve resolves a DNS address
func (d StatDNSResolver) Resolve(name string) (net.IP, error) {
	result, err := RunDNSLookup(name)
	if err != nil {
		return nil, err
	}
	for _, answer := range result.Answer {
		if answer.Type != "CNAME" {
			answer := answer.RData
			return net.ParseIP(answer), err
		}
	}
	return nil, errors.New("No records exist")
}

func main() {
	// Create a SOCKS5 server
	var resolver StatDNSResolver
	conf := &socks5.Config{Resolver: resolver}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", "127.0.0.1:9001"); err != nil {
		panic(err)
	}
}
