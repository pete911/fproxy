package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var httpClient = http.Client{
	Timeout: 10 * time.Second,
}

func init() {
	log.SetFlags(0)
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

	log.Printf("* %s %s", r.URL.Host, r.URL.Port())
	log.Printf("> request: %s %s %s", r.Method, r.URL.Path, r.Proto)
	for key, values := range r.Header {
		log.Printf("> %s: %s", key, strings.Join(values, ", "))
	}
	log.Print(">")

	proxyRequest := http.Request{
		Method: r.Method,
		URL: r.URL,
		Header: r.Header,
		Body: r.Body,
		ContentLength: r.ContentLength,
		Close: r.Close,
	}

	proxyResponse, err := httpClient.Do(&proxyRequest)
	if err != nil {
		log.Printf("proxy request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - proxy error"))
		return
	}
	defer proxyResponse.Body.Close()

	// write headers
	log.Printf("< response: %s", proxyResponse.Status)
	for key, values := range proxyResponse.Header {
		log.Printf("< %s: %s", key, strings.Join(values, ", "))
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(proxyResponse.StatusCode)
	if _, err := io.Copy(w, proxyResponse.Body); err != nil {
		log.Printf("io copy proxy response: %v", err)
	}
	log.Print("<")
}
