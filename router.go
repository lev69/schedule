package main

import (
	"net/http"

	schedule "example.com/schedule/lib"

	httpSwagger "github.com/swaggo/http-swagger"
)

func initRouter() {
	http.HandleFunc("/user", schedule.UserHandler)
	http.HandleFunc("/meeting", schedule.MeetingHandler)
	http.HandleFunc("/response", schedule.ResponseHandler)
	http.HandleFunc("/user_meetings", schedule.UserMeetingsHandler)

	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

}
