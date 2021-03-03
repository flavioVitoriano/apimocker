package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

// Mocked holds the code and payload for the endpoint
type Mocked struct {
	Code    int
	Type    string
	Payload string
}

// Verb holds the mocked structures
type Verb map[string]Mocked

// Endpoints holds the verbs
type Endpoints map[string]Verb

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file>\n", os.Args[0])
		os.Exit(1)
	}

	endpoints := parseFile(os.Args[1])

	startServer(endpoints)

}

// loads the file and return the expected values
func parseFile(filePath string) Endpoints {
	mockFile := os.Args[1]
	file, err := ioutil.ReadFile(mockFile)
	var endpoints Endpoints

	if err != nil {
		log.Fatal(err)
	}

	yaml.Unmarshal(file, &endpoints)

	return endpoints
}

// create handler based on endpoints
func createHandler(endpoints Endpoints) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		message := []byte(endpoints[request.URL.Path][request.Method].Payload)
		proxy := endpoints[request.URL.Path][request.Method].Payload
		mockType := endpoints[request.URL.Path][request.Method].Type
		code := endpoints[request.URL.Path][request.Method].Code

		// If the method is POST and the payload is empty just echo the
		// payload back

		if request.Method == "POST" && len(message) == 0 {
			body, _ := ioutil.ReadAll(request.Body)
			defer request.Body.Close()
			message = body
		}

		writer.WriteHeader(code)

		// check mock type
		if mockType == "mock" {
			_, err := writer.Write(message)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			resp, err := http.Get(proxy)

			if err != nil {
				log.Fatal(err)
			}

			// read body from response
			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				log.Fatal(err)
			}

			writer.Write(body)
		}
	}
}

// start server
func startServer(endpoints Endpoints) {
	for endpoint := range endpoints {
		http.HandleFunc(endpoint, createHandler(endpoints))
	}

	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		log.Fatal(err)
	}
}
