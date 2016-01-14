package response

import (
	"net/http"
)

type HttpResponse struct {
	Cookies []*http.Cookie
	Body    []byte
	Header  http.Header
}

func GetResponse(response *http.Response) *HttpResponse {
	var http_resp HttpResponse
	http_resp.Cookies = response.Cookies()
	http_resp.Header = response.Header

	return &http_resp
}
