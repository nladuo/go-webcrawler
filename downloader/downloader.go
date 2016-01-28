package downloader

import (
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"net/http"
	netUrl "net/url"
)

const (
	RETRY_TIMES int = 10
)

var (
	request http.Request
)

func Download(tag string, task model.Task) *model.Result {
	var err error
	var resp *http.Response
	var result model.Result
	var retry_times = 0
	log.Println(tag, "Start Download: ", task.Url)
REDOWNLOAD:

	if proxyStr := task.GetProxy(); len(proxyStr) == 0 {
		resp, err = dowloadDirect(task.Url)
	} else {
		resp, err = dowloadWithProxy(task.Url, proxyStr)
	}
	if err != nil {
		if retry_times > RETRY_TIMES {
			return nil
		}
		retry_times++
		goto REDOWNLOAD
	}
	//result.Response
	result.Identifier = task.Identifier
	result.UserData = task.UserData
	result.Response, result.Err = model.GetResponse(resp)
	log.Println(tag, "Download Success: ", task.Url)
	return &result
}

func dowloadDirect(url string) (*http.Response, error) {
	return http.Get(url)
}

func dowloadWithProxy(url, proxyStr string) (*http.Response, error) {
	request, _ := http.NewRequest("GET", url, nil)
	proxy, err := netUrl.Parse(proxyStr)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	return client.Do(request)
}
