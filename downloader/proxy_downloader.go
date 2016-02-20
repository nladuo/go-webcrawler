package downloader

import (
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"net/http"
)

type ProxyDownloader struct {
	generator  model.ProxyGenerator
	retryTimes int
}

func NewProxyDownloader(generator model.ProxyGenerator) *ProxyDownloader {
	var downloader ProxyDownloader
	downloader.retryTimes = 10
	downloader.generator = generator
	return &downloader
}

func (this *ProxyDownloader) SetProxyGenerator(generator model.ProxyGenerator) {
	this.generator = generator
}

func (this *ProxyDownloader) Download(tag string, task model.Task) *model.Result {
	var err error
	var resp *http.Response
	var result model.Result
	var retry_times = 0
	log.Println(tag, "Start Download: ", task.Url)
REDOWNLOAD:

	if proxy := this.generator.GetProxy(); len(proxy.IP) == 0 {
		log.Println(tag, "You haven't set proxy.")
		return nil
	} else {
		resp, err = dowloadByProxy(task.Url, &proxy)
	}

	if err != nil {
		if retry_times > this.retryTimes {
			log.Println(tag, "Download Failed: ", task.Url, "Error:", err.Error())
			return nil
		}
		retry_times++
		goto REDOWNLOAD
	}

	result.Identifier = task.Identifier
	result.Url = task.Url
	result.UserData = task.UserData
	result.Response, result.Err = model.GetResponse(resp)
	log.Println(tag, "Download Success: ", task.Url)
	return &result

}

func (this *ProxyDownloader) SetRetryTimes(times int) {
	this.retryTimes = times
}
