// The schedule program is a simple service providing meetings interface.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "example.com/schedule/docs"
)

// @title        Schedule API
// @version      0.9
// @description  Schedule is simple calendare service
// @license.name WTFPL
// @host         localhost:8000
func main() {
	port := flag.Uint("p", 8000, "Listen on the port")
	address := flag.String("a", "localhost", "Bind to the local address")

	initRouter()
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil))
}
