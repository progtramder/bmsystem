package main

import (
	"html/template"
	"net/http"
	"log"
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

	if checkAuth(r) {
		http.FileServer(http.Dir(self)).ServeHTTP(w, r)
		return
	}

	code := r.FormValue("code")
	openId := GetOpenId(code)
	if openId != "" {
		http.FileServer(http.Dir(self)).ServeHTTP(w, r)
		return
	}

	adminpage := systembasePath + "/webroot/html/admin.html"
	t, err := template.ParseFiles(adminpage)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Execute(w, "/report/" + r.URL.Path)
}