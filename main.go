// The schedule program is a simple service providing meetings interface.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	schedule "example.com/schedule/lib"
)

func main() {
	port := flag.Uint("p", 8000, "Listen on the port")
	address := flag.String("a", "localhost", "Bind to the local address")

	http.HandleFunc("/user", schedule.UserHandler)
	http.HandleFunc("/meeting", schedule.MeetingHandler)
	http.HandleFunc("/response", schedule.ResponseHandler)
	http.HandleFunc("/user_meetings", schedule.UserMeetingsHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil))
}