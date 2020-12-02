package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type PrintBody func(headerContentType, bodyContentType string) bool

type Headers map[string]string

func (h Headers) PrettyString(prefix string) string {

	var headers []string
	for k, v := range h {
		headers = append(headers, fmt.Sprintf("%s%s: %s", prefix, k, v))
	}
	return strings.Join(headers, "\n")
}

type Request struct {
	Method      string
	URL         string
	Proto       string
	Host        string
	Headers     Headers
	ContentType string
	Body        []byte
}

func (r Request) PrettyString(printBody PrintBody) string {

	var body = fmt.Sprintf("not printing %s content", r.ContentType)
	if printBody(r.Headers["Content-Type"], r.ContentType) {
		body = string(r.Body)
	}
	template := `> %s %s %s
> Host: %s
%s
>
* %s
%s
`
	return fmt.Sprintf(template, r.Method, r.URL, r.Proto, r.Host, r.Headers.PrettyString("> "), r.ContentType, body)
}

type Response struct {
	Proto       string
	Status      string
	Headers     Headers
	ContentType string
	Body        []byte
}

func (r Response) PrettyString(printBody PrintBody) string {

	var body = fmt.Sprintf("not printing %s content", r.ContentType)
	if printBody(r.Headers["Content-Type"], r.ContentType) {
		body = string(r.Body)
	}
	template := `< %s %s
%s
<
* %s
%s
`
	return fmt.Sprintf(template, r.Status, r.Proto, r.Headers.PrettyString("< "), r.ContentType, body)
}

type Proxy struct {
	client *http.Client
}

func NewProxy(client *http.Client) Proxy {
	return Proxy{client: client}
}

func (p Proxy) Forward(w http.ResponseWriter, r *http.Request) (Request, Response, error) {

	rBody, requestBody, err := copyBody(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return Request{}, Response{}, fmt.Errorf("copy request body: %+v", err)
	}

	proxyRequest := http.Request{
		Method:        r.Method,
		URL:           r.URL,
		Header:        r.Header,
		Body:          rBody,
		ContentLength: r.ContentLength,
		Close:         r.Close,
	}

	response, err := p.client.Do(&proxyRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return Request{}, Response{}, fmt.Errorf("proxy request: %+v", err)
	}

	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return Request{}, Response{}, fmt.Errorf("read response body: %+v", err)
	}
	writeResponse(w, response.StatusCode, response.Header, responseBody)

	return Request{
			Method:      r.Method,
			URL:         r.URL.String(),
			Proto:       r.Proto,
			Host:        r.Host,
			Headers:     toHeaders(r.Header),
			ContentType: http.DetectContentType(requestBody),
			Body:        requestBody,
		},
		Response{
			Proto:       response.Proto,
			Status:      response.Status,
			Headers:     toHeaders(response.Header),
			ContentType: http.DetectContentType(responseBody),
			Body:        responseBody,
		}, nil
}

func writeResponse(w http.ResponseWriter, statusCode int, header http.Header, body []byte) {

	for key, values := range header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		log.Printf("write response body: %v", err)
	}
}

func copyBody(r io.ReadCloser) (io.ReadCloser, []byte, error) {

	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(b)), b, nil
}

func toHeaders(header http.Header) map[string]string {

	headers := make(map[string]string)
	for key, values := range header {
		headers[key] = strings.Join(values, ",")
	}
	return headers
}
