package model

import (
	"io/ioutil"
	"net/http"
)

type HttpResponse struct {
	Cookies       []*http.Cookie
	Body          []byte
	Header        http.Header
	StatusCode    int
	ContentLength int64
}

func GetResponse(response *http.Response) (*HttpResponse, error) {
	var http_resp HttpResponse
	var err error
	http_resp.Body, err = ioutil.ReadAll(response.Body)
	http_resp.Cookies = response.Cookies()
	http_resp.Header = response.Header
	http_resp.StatusCode = response.StatusCode
	http_resp.ContentLength = response.ContentLength

	return &http_resp, err
}
