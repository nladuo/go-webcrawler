package downloader

import (
	//"errors"
	"github.com/nladuo/go-webcrawler/model"
	//"log"
	"net/http"
	netUrl "net/url"
	"time"
)

var (
	proxyTimeOut time.Duration = time.Duration(0)
)

type Downloader interface {
	Download(tag string, task model.Task) *model.Result
	SetRetryTimes(times int)
}

func dowloadDirect(url string) (*http.Response, error) {
	return http.Get(url)
}

func dowloadByProxy(url string, proxy *model.Proxy) (*http.Response, error) {
	request, _ := http.NewRequest("GET", url, nil)
	proxyStr := "http://" + proxy.IP + ":" + proxy.Port
	proxyUrl, err := netUrl.Parse(proxyStr)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
	//if has time out
	if !(proxyTimeOut == time.Duration(0)) {
		client.Timeout = proxyTimeOut
	}
	return client.Do(request)
}

func SetProxyTimeOut(timeout time.Duration) {
	proxyTimeOut = timeout
}
