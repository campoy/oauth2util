// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program runs a simple HTTP server and every request is
// forwarded for oauth2 authentication to request access to the
// Google calendar's API for the visiting user.
//
// Then those credentials are used to obtain a list of up to
// ten events from the user's primary calendar and the summaries
// of those events are printed.
package main

import (
	"fmt"
	"log"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/calendar/v3"
	"github.com/campoy/oauth2util"
)

// config should be modified in your code setting the good values for
// ClientId, ClientSecret, RedirectURL, and Scope.
// Note that RedirectURL should point to the path oauth2util.CallbackURL
// on your server.
var config = &oauth.Config{
	ClientId:     "your-client-id",
	ClientSecret: "your-client-secret",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
	RedirectURL:  "http://localhost:8080" + oauth2util.CallbackURL,
	Scope:        calendar.CalendarScope,
}

func main() {
	// We could ignore the error, the only possible error is a wrong
	// configuration.
	err := oauth2util.HandleFunc("/", eventsHandler, config)
	if err != nil {
		panic(err)
	}
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

// eventsHandler is executed when the authentication mechanism redirects
// from the authentication server to the application server.
func eventsHandler(w http.ResponseWriter, r *http.Request) {
	// Obtain a new authenticated http.Client for the requested config.
	client, err := oauth2util.Client(r, nil, config)
	if err != nil {
		http.Error(w, "oauth2 client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new calendarAPI client, using the authenticated client.
	cal, err := calendar.New(client)
	if err != nil {
		http.Error(w, "create calendar service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Obtain up to 10 events, older than may 28th, from the primary calendar of the user.
	evts, err := cal.Events.List("primary").
		MaxResults(10).
		TimeMin("2013-05-28T00:00:00-08:00").
		Do()
	if err != nil {
		http.Error(w, "get calendar events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Print the summaries of all the events onto the HTTP response.
	for _, evt := range evts.Items {
		fmt.Fprintf(w, "<p>%v</p>\n", evt.Summary)
	}
}
