package downloader

import (
	"errors"
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"net/http"
	netUrl "net/url"
	//"time"
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
	proxyStr := proxy.IP + ":" + proxy.Port
	switch proxy.Type {
	case model.TYPE_HTTP:
		proxyStr = "http://" + proxyStr
	case model.TYPE_HTTPS:
		proxyStr = "https://" + proxyStr
	default:
		log.Println("You set a proxy of which type is neither HTTP nor HTTPS")
		return nil, errors.New(model.ErrProxyTypeNotExist)
	}
	proxyUrl, err := netUrl.Parse(proxyStr)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
	//client.Timeout = 10 * time.Second
	return client.Do(request)
}
