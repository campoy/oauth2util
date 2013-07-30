// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The oauth2util package provides support for oauth2-authenticated
// HTTP handlers. Specially when multiple authentication configurations
// are needed.
package oauth2util

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"code.google.com/p/goauth2/oauth"
)

// CallbackURL is the path that the RedirectURL field in the config file should have.
// If they differ, Handle and HandleFunc will return an error.
const CallbackURL = "/oauth2callback"

var (
	mutex    sync.RWMutex
	handlers = make(map[string]http.HandlerFunc)
)

func init() {
	http.HandleFunc(CallbackURL, callback)
}

// Handle register the passed http.Handler to be executed when
// a request matches url. The handler will receive the oauth2
// code corresponding to the passed oauth.Config as form values.
func Handle(url string, h http.Handler, cfg *oauth.Config) error {
	return HandleFunc(url, h.ServeHTTP, cfg)
}

// Handle register the passed http.HandleFunc to be executed when
// a request matches url. The HandleFunc will receive the oauth2
// code corresponding to the passed oauth.Config as form values.
func HandleFunc(path string, h http.HandlerFunc, cfg *oauth.Config) error {
	mutex.Lock()
	defer mutex.Unlock()

	u, err := url.Parse(cfg.RedirectURL)
	if err != nil {
		return fmt.Errorf("bad redirect URL: %v", err)
	}
	if u.Path != CallbackURL {
		return fmt.Errorf("RedirectURL has to point to %q, it points to %q", CallbackURL, u.Path)
	}

	handlers[path] = h
	rh := http.RedirectHandler(cfg.AuthCodeURL(path), http.StatusFound)
	http.Handle(path, rh)
	return nil
}

// Client creates an authenticated http.Client given an http.Request
// containing an oauth code.
func Client(r *http.Request, transport http.RoundTripper, cfg *oauth.Config) (*http.Client, error) {
	if errp := r.FormValue("error"); len(errp) > 0 {
		return nil, fmt.Errorf("error in oauth2 response: %q", errp)
	}
	t := &oauth.Transport{
		Config:    cfg,
		Transport: transport,
	}
	_, err := t.Exchange(r.FormValue("code"))
	if err != nil {
		return nil, err
	}
	return t.Client(), nil
}

// callback handles the response from the authentication server and redirects
// it to the corresponding handler.
func callback(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	state := r.FormValue("state")
	h, ok := handlers[state]
	if !ok {
		http.Error(w, "unexpected state "+state, http.StatusBadRequest)
		return
	}
	h(w, r)
}
