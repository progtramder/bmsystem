package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type router struct {
}

func (this *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	urlString := "http://localhost:82"
	if strings.Contains(path, "/.well-known/acme-challenge") {
		urlString = "http://localhost:8080"
	}

	remote, _ := url.Parse(urlString)
	proxy := httputil.NewSingleHostReverseProxy(remote)
	r.Host = remote.Host
	proxy.ServeHTTP(w, r)
}
