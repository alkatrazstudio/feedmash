// SPDX-License-Identifier: AGPL-3.0-only
// 🄯 2021, Alexey Parfenov <zxed@alkatrazstudio.net>

package src

import (
	"feedmash/util"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func serverHandler(w http.ResponseWriter, outXml string) {
	if outXml == "" {
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/atom+xml")
	w.Header().Add("Content-Length", strconv.Itoa(len(outXml)))

	_, err := fmt.Fprint(w, outXml)
	if err != nil {
		util.LogWarn(err)
	}
}

func runServer(addr string, stop chan bool, stopped chan bool, outXmlChan chan string) {
	outXml := ""

	go func() {
		for {
			outXml = <-outXmlChan
		}
	}()

	srv := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			serverHandler(w, outXml)
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 12,
	}

	go func() {
		util.LogInfo("Starting server at http://" + addr)
		err := srv.ListenAndServe()
		if err != nil {
			util.LogWarn(err)
			stopped <- true
		}
	}()
	<-stop
	err := srv.Close()
	if err != nil {
		util.LogWarn(err)
	}
	stopped <- true
}
