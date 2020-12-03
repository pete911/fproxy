package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	proxy  = NewProxy(&http.Client{Timeout: 10 * time.Second})
	port   int
	tlsKey string
	tlsCrt string
	output *os.File
)

func init() {

	f, err := ParseFlags()
	if err != nil {
		Errorf("cannot parse flags: %v", err)
		os.Exit(1)
	}

	Logf("flags: %+v", f)
	Silent = f.Silent
	tlsCrt = f.TLSCrt
	tlsKey = f.TLSKey
	port = f.Port

	if f.OutputFile == "" {
		output = os.Stdout
		return
	}

	outputFile, err := os.OpenFile(f.OutputFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		Errorf("cannot open %s output file: %v", f.OutputFile, err)
		os.Exit(1)
	}
	output = outputFile
}

func main() {

	handler := http.DefaultServeMux
	handler.HandleFunc("/", handleFunc)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if tlsCrt == "" && tlsKey == "" {
		Errorf("listen and serve: %v", s.ListenAndServe())
	} else {
		Errorf("listen and serve TLS: %v", s.ListenAndServeTLS(tlsCrt, tlsKey))
		os.Exit(1)
	}
}

func handleFunc(w http.ResponseWriter, r *http.Request) {

	request, response, err := proxy.Forward(w, r)
	if err != nil {
		Errorf("proxy forward: %v", err)
		return
	}

	now := time.Now().Format("20060102-15:04:05")
	fmt.Fprintf(output, "[%s %s]\n", now, r.URL)
	fmt.Fprint(output, request.PrettyString(printRequestBody))
	fmt.Fprint(output, response.PrettyString(printResponseBody))
}

func printRequestBody(_, bodyContentType string) bool {

	for _, allowedContent := range []string{"text", "html", "json", "xml"} {
		if strings.Contains(bodyContentType, allowedContent) {
			return true
		}
	}
	return false
}

func printResponseBody(_, bodyContentType string) bool {

	for _, allowedContent := range []string{"text", "html", "json", "xml"} {
		if strings.Contains(bodyContentType, allowedContent) {
			return true
		}
	}
	return false
}
