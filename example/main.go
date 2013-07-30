package main

import (
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/calendar/v3"
	"github.com/campoy/oauth2util"
)

var config = &oauth.Config{
	ClientId:     "your-client-id",
	ClientSecret: "your-client-secret",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
	RedirectURL:  "http://localhost:8080" + oauth2util.CallbackURL,
	Scope:        calendar.CalendarScope,
}

func main() {
	err := oauth2util.HandleFunc("/", eventsHandler, config)
	if err != nil {
		panic(err)
	}
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	client, err := oauth2util.Client(r, nil, config)
	if err != nil {
		http.Error(w, "oauth2 client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cal, err := calendar.New(client)
	if err != nil {
		http.Error(w, "create calendar service: "+err.Error(), http.StatusInternalServerError)
		return
	}
	evts, err := cal.Events.List("primary").
		MaxResults(10).
		TimeMin("2013-05-28T00:00:00-08:00").
		Do()

	if err != nil {
		http.Error(w, "get calendar events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, evt := range evts.Items {
		fmt.Fprintf(w, "<p>%v</p>\n", evt.Summary)
	}
}
