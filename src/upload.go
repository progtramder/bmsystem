package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type tokens struct {
	sync.Mutex
	tokens map[string]bool
}

func (t *tokens) put(token string) {
	t.Lock()
	t.tokens[token] = true
	t.Unlock()
}

func (t *tokens) get(token string) bool {
	t.Lock()
	_, ok := t.tokens[token]
	t.Unlock()
	return ok
}

var tokenPool = tokens{tokens: map[string]bool{}}

func NewToken() string {
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(time.Now().Unix(), 10))
	token := fmt.Sprintf("%x", h.Sum(nil))
	return token
}

func developResponse(res http.ResponseWriter, msg string) {
	resp := struct {
		Token string
		Msg   string
	}{}

	token := NewToken()
	t, err := template.ParseFiles(systembasePath + "/webroot/html/upload.html")
	if err != nil {
		panic(err)
	}

	resp.Token = token
	resp.Msg = msg
	tokenPool.put(token)
	t.Execute(res, resp)
}

func savePoster(fh *multipart.FileHeader) error {
	fileName := filepath.Base(fh.Filename)
	root := "/webroot/image/"

	file, err := fh.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(systembasePath+root+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := io.Copy(f, file)
	//if the size of new file is less than the existed file, truncate is must,
	f.Truncate(n)
	return err
}

func saveUpload(fh *multipart.FileHeader) error {
	fileName := filepath.Base(fh.Filename)
	s := strings.Split(fileName, ".")
	if len(s) == 0 {
		return errors.New("unknown file type of " + fileName)
	}
	suffix := s[len(s)-1]
	var root string
	switch strings.ToLower(suffix) {
	case "js":
		root = "/webroot/js/"
	case "html":
		root = "/webroot/html/"
	case "css":
		root = "/webroot/css/"
	case "jpeg", "jpg", "png":
		root = "/webroot/image/"
	case "txt":
		root = "/webroot/"
	case "yaml":
		if fileName != "event.yaml" {
			return errors.New("file name must be event.yaml")
		}
		root = "/"
	default:
		return errors.New("unsupported file type")
	}

	file, err := fh.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(systembasePath+root+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := io.Copy(f, file)
	//if the size of new file is less than the existed file, truncate is must,
	f.Truncate(n)
	return err
}

func saveToAlbum(fh *multipart.FileHeader) error {
	fileName := filepath.Base(fh.Filename)
	s := strings.Split(fileName, ".")
	if len(s) == 0 {
		return errors.New("unknown file type of " + fileName)
	}
	suffix := s[len(s)-1]
	var root string
	switch strings.ToLower(suffix) {
	case "jpeg", "jpg", "png", "mp4":
		root = "/webroot/album/"
	default:
		return errors.New("unsupported file type")
	}

	file, err := fh.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(systembasePath+root+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := io.Copy(f, file)
	//if the size of new file is less than the existed file, truncate is must,
	f.Truncate(n)
	return err
}
