package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	proxy  = NewProxy(&http.Client{Timeout: 10 * time.Second})
	output io.Writer
)

func init() {
	log.SetFlags(0)
	output = os.Stdout
}

func main() {

	handler := http.DefaultServeMux
	handler.HandleFunc("/", handleFunc)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatalf("listen and serve: %v", s.ListenAndServe())
}

func handleFunc(w http.ResponseWriter, r *http.Request) {

	request, response, err := proxy.Forward(w, r)
	if err != nil {
		log.Print(err)
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
