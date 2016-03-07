package downloader

import (
	//"errors"
	"github.com/nladuo/go-webcrawler/model"
	//"log"
	"net/http"
	netUrl "net/url"
	"time"
)

const (
	ErrProxyNotSet = "you haven't set proxy"
)

var (
	//default timeout
	proxyTimeOut time.Duration = 0 * time.Second
)

type Downloader interface {
	Download(tag string, task model.Task) model.Result
	SetRetryTimes(times int)
}

func dowloadDirect(url string) (*http.Response, error) {
	return http.Get(url)
}

func dowloadWithProxy(url string, proxy model.Proxy) (*http.Response, error) {
	request, _ := http.NewRequest("GET", url, nil)
	proxyStr := "http://" + proxy.IP + ":" + proxy.Port
	proxyUrl, err := netUrl.Parse(proxyStr)
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	if proxyTimeOut != 0*time.Second {
		client.Timeout = proxyTimeOut
	}

	return client.Do(request)
}

func SetProxyTimeOut(timeout time.Duration) {
	proxyTimeOut = timeout
}
