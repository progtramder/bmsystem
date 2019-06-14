package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

func handleBM(w http.ResponseWriter, r *http.Request) {
	event := r.FormValue("event")
	bmEvent := bmEventList.GetEvent(event)
	if bmEvent == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	t, err := template.ParseFiles(systembasePath + "/webroot/html/render.html")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	htmlInfo := struct {
		WXCode string
		Event  string
	}{code, event}
	err = t.Execute(w, htmlInfo)
}

func handleEventProfile(w http.ResponseWriter, r *http.Request) {
	event := r.FormValue("event")
	bmEvent := bmEventList.GetEvent(event)
	if bmEvent == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	openId := GetOpenId(code)
	if openId == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	poster := ""
	if bmEvent.poster != "" {
		poster = fmt.Sprintf("https://%s/image/%s", r.Host, bmEvent.poster)
	}
	page := struct {
		OpenId string      `json:"openid"`
		Poster string      `json:"poster"`
		Form   []Component `json:"form"`
	}{
		openId,
		poster,
		bmEvent.form,
	}

	b, _ := json.Marshal(&page)
	w.Write(b)
}

func handleSubmitBM(w http.ResponseWriter, r *http.Request) {
	event := r.FormValue("event")
	bmEvent := bmEventList.GetEvent(event)
	if bmEvent == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	openId := r.FormValue("openid")
	if openId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	info := bminfo{}
	info.Load(data)
	errCode := bmEvent.put(openId, info)
	if errCode == errSuccess {
		bmEvent.serialize(openId, info)
	}
	w.Write([]byte(fmt.Sprintf(`{"errCode":%d,"errMsg":"%s"}`, errCode, Reason(errCode))))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	event := r.FormValue("event")
	bmEvent := bmEventList.GetEvent(event)
	if bmEvent == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bmEvent.RLock()
	defer bmEvent.RUnlock()

	type _session struct {
		Desc    string `json:"description"`
		Limit   int    `json:"limit"`
		Number  int    `json:"number"`
		Expired bool   `json:"expired"`
	}
	status := struct {
		Started  bool       `json:"started"`
		Sessions []_session `json:"sessions"`
	}{
		Started: bmEvent.started,
	}

	for _, v := range bmEvent.sessions {
		s := _session{
			v.Desc,
			v.Limit,
			v.number,
			v.Expired(),
		}
		status.Sessions = append(status.Sessions, s)
	}

	b, _ := json.Marshal(&status)
	w.Write(b)
}

func handleRegisterInfo(w http.ResponseWriter, r *http.Request) {
	event := r.FormValue("event")
	bmEvent := bmEventList.GetEvent(event)
	if bmEvent == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	openId := r.FormValue("openid")
	if openId == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bmEvent.RLock()
	defer bmEvent.RUnlock()
	info, _ := bmEvent.has(openId)
	w.Write([]byte(info.Dump()))
}

// Admin handlers
func checkAuth(r *http.Request) bool {
	h := md5.New()
	io.WriteString(h, privateData.AdminPassword)
	pass := fmt.Sprintf("%x", h.Sum(nil))
	for _, v := range r.Cookies() {
		if v.Name == "admin_password" && v.Value == pass {
			return true
		}
	}

	return false
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	adminpage := systembasePath + "/webroot/html/login.html"

	if checkAuth(r) {
		adminpage = systembasePath + "/webroot/html/events.html"
	}

	t, err := template.ParseFiles(adminpage)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
}

// Admin handlers
func handleStartBaoming(w http.ResponseWriter, r *http.Request) {
	event := r.FormValue("event")
	bmEvent := bmEventList.GetEvent(event)
	if bmEvent == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if checkAuth(r) {
		bmEvent.Start()
	}
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" || !tokenPool.get(token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := bmEventList.Reset()
	if err != nil {
		ColorRed("Fail to reset: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func handleDevelop(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if checkAuth(r) {
			developResponse(w, "")
			return
		}

		adminpage := systembasePath + "/webroot/html/develop.html"

		t, err := template.ParseFiles(adminpage)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		t.Execute(w, nil)
	} else {
		token := r.PostFormValue("token")
		if token == "" || !tokenPool.get(token) {
			developResponse(w, fmt.Sprintf("未授权"))
			return
		}

		_, _, err := r.FormFile("uploadfile")
		if err != nil {
			developResponse(w, fmt.Sprintf("上传失败 : %v", err))
			return
		}

		fhs := r.MultipartForm.File["uploadfile"]
		for _, v := range fhs {
			if err := saveUpload(v); err != nil {
				developResponse(w, fmt.Sprintf("上传失败 : %v", err))
				return
			}
		}

		developResponse(w, fmt.Sprintf("上传成功"))
	}
}

func handleSaveAlbum(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		token := r.FormValue("token")
		if token == "" || !tokenPool.get(token) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _, err := r.FormFile("save-album")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("上传失败 : %v", err)))
			return
		}

		fhs := r.MultipartForm.File["save-album"]
		fh := fhs[0]
		fileName := filepath.Base(fh.Filename)
		if err := saveToAlbum(fh); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte( fmt.Sprintf("上传失败 : %v", err)))
			return
		}
		w.Write([]byte("album/" + fileName))

	} else if r.Method == "GET" {
		code := r.FormValue("code")
		openId := GetOpenId(code)
		if openId == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h := md5.New()
		io.WriteString(h, strconv.FormatInt(time.Now().Unix(), 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		tokenPool.put(token)
		w.Write([]byte(fmt.Sprintf(`{"token":"%s"}`, token)))
	}
}

func handlGetEvents(w http.ResponseWriter, _ *http.Request) {
	events := struct {
		Data []string `json:"data"`
	}{}

	bmEventList.RLock()
	defer bmEventList.RUnlock()
	for _, v := range bmEventList.events {
		events.Data = append(events.Data, v.name)
	}

	b, _ := json.Marshal(&events)
	w.Write(b)
}
