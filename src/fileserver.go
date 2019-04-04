package main

import (
	"net/http"
	"strings"
)

type fileHandler string

func FileServer(dir string) http.Handler {
	return fileHandler(dir)
}

func (self fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !checkAuth(r) && (path == "" || strings.HasSuffix(path, "/") || strings.Contains(path, "html/")) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	http.FileServer(http.Dir(self)).ServeHTTP(w, r)
}

type reportHandler string

func ReportServer(dir string) http.Handler {
	return reportHandler(dir)
}

func (self reportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	http.FileServer(http.Dir(self)).ServeHTTP(w, r)
}
