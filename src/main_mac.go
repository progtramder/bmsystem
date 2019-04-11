// +build darwin

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	"ziphttp"
)

var (
	systembasePath string
)

func main() {
	systembasePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	err := InitPrivate()
	if err != nil {
		log.Fatal(err)
	}

	err = bmEventList.Reset()
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		fmt.Println("App server start listening on port:443 ...")
		mux := http.ServeMux{}
		mux.Handle("/", FileServer(systembasePath+"/webroot"))
		mux.Handle("/report/", http.StripPrefix("/report/", ReportServer(systembasePath+"/report")))
		mux.HandleFunc("/baoming", handleBM)
		mux.HandleFunc("/event-profile", handleEventProfile)
		mux.HandleFunc("/submit-baoming", handleSubmitBM)
		mux.HandleFunc("/status", handleStatus)
		mux.HandleFunc("/register-info", handleRegisterInfo)
		mux.HandleFunc("/start-baoming", handleStartBaoming)
		mux.HandleFunc("/admin", handleAdmin)
		mux.HandleFunc("/develop", handleDevelop)
		mux.HandleFunc("/reset", handleReset)
		mux.HandleFunc("/get-events", handlGetEvents)
		srv := &http.Server{
			Addr:        ":443",
			ReadTimeout: 5 * time.Second,
			Handler:     &mux,
		}

		wg.Done()
		log.Fatal(srv.ListenAndServe())

	}()

	//wait for server starting
	wg.Wait()
	fmt.Println("Done.")

	ziphttp.CmdLineLoop(prompt, func(input string) int {
		handler, ok := CmdLineHandler[input]
		if ok {
			return handler.Handle()
		}

		return Continue()
	})
}
