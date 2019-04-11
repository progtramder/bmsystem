// +build !darwin

package main

import (
	"crypto/tls"
	"fmt"
	"github.com/mholt/certmagic"
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

	//Starting http router, routing to acme challenge server and app iis
	go func() {
		fmt.Println("Http router start listening on port:81 ...")
		log.Fatal(http.ListenAndServe(":81", &router{}))
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		time.Sleep(time.Millisecond * 100)
		certmagic.Agreed = true
		certmagic.Email = "wolf_wml@163.com"
		certmagic.CA = certmagic.LetsEncryptProductionCA
		certmagic.AltHTTPPort = 8080
		magic := certmagic.New(certmagic.Config{})
		err := magic.Manage([]string{privateData.Domain})
		if err != nil {
			log.Fatal(err)
		}

		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			magic.HandleHTTPChallenge(w, r)
		})
		httpSrv := &http.Server{
			Addr:        ":8080",
			ReadTimeout: 5 * time.Second,
			Handler:     httpHandler,
		}
		go func() {
			fmt.Println("Acme challenge server start listening on port:8080 ...")
			log.Fatal(httpSrv.ListenAndServe())
		}()

		time.Sleep(time.Millisecond * 100)
		fmt.Println("App server start listening on port:443 ...")
		httpsLn, err := tls.Listen("tcp", ":443", magic.TLSConfig())
		if err != nil {
			log.Fatal(err)
		}

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
		tlsSrv := &http.Server{
			ReadTimeout: 5 * time.Second,
			Handler:     &mux,
		}

		wg.Done()
		log.Fatal(tlsSrv.Serve(httpsLn))
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
