package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	flags  Flags
	proxy  = NewProxy(&http.Client{Timeout: 10 * time.Second})
	output io.Writer
)

func init() {

	f, err := ParseFlags()
	if err != nil {
		Errorf("cannot parse flags: %v", err)
		os.Exit(1)
	}
	flags = f
	Verbose = flags.Verbose
	Logf("flags: %+v", flags)

	// TODO - set output to either stdout or file based on flag f.OuputFile == ""
	output = os.Stdout
}

func main() {

	handler := http.DefaultServeMux
	handler.HandleFunc("/", handleFunc)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", flags.Port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if flags.TLSCrt == "" && flags.TLSKey == "" {
		Errorf("listen and serve: %v", s.ListenAndServe())
	} else {
		Errorf("listen and serve TLS: %v", s.ListenAndServeTLS(flags.TLSCrt, flags.TLSKey))
		os.Exit(1)
	}
}

func handleFunc(w http.ResponseWriter, r *http.Request) {

	request, response, err := proxy.Forward(w, r)
	if err != nil {
		Errorf("proxy forward: %v", err)
		return
	}
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
