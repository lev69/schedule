// The schedule program is a simple service providing meetings interface.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Uint("p", 8000, "Listen on the port")
	address := flag.String("a", "localhost", "Bind to the local address")

	initRouter()
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil))
}
