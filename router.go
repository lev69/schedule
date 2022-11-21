package main

import (
	"net/http"

	schedule "github.com/lev69/schedule/lib"

	httpSwagger "github.com/swaggo/http-swagger"
)

func initRouter() {
	http.HandleFunc("/user", schedule.UserHandler)
	http.HandleFunc("/meeting", schedule.MeetingHandler)
	http.HandleFunc("/response", schedule.ResponseHandler)
	http.HandleFunc("/user_meetings", schedule.UserMeetingsHandler)
	http.HandleFunc("/find_free_time", schedule.FindFreeTimeHandler)

	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}
