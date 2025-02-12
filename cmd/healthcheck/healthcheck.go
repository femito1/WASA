/*
Healthcheck is a simple program that sends an HTTP request to the local host (self) to a configured port number.
It's used in environments where you need a simple probe for health checks (e.g., an empty container in docker).
The probe URL is http://localhost:3000/liveness. Only the port can be changed.

Usage:

    healthcheck [flags]

The flags are:

    -port <1-65535>
        Change the port where the request is sent.

Return values (exit codes):

    0
        The request was successful (HTTP 200 or HTTP 204)

    > 0
        The request was not successful (connection error or unexpected HTTP status code)
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Create a logger that writes to standard error without additional prefixes or flags.
	logger := log.New(os.Stderr, "", 0)

	// Define and parse the port flag.
	var port = flag.Int("port", 3000, "HTTP port for healthcheck")
	flag.Parse()

	// Perform the health check request.
	res, err := http.Get(fmt.Sprintf("http://localhost:%d/liveness", *port))
	if err != nil {
		logger.Println(err.Error())
		return
	}
	defer res.Body.Close()

	// Check the status code.
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		logger.Println("Healthcheck request not OK:", res.Status)
		return
	}
}
