package downloader

import (
	"errors"
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"net/http"
)

type ProxyDownloader struct {
	generator model.ProxyGenerator
	//retryTimes int
}

func NewProxyDownloader(generator model.ProxyGenerator) *ProxyDownloader {
	var downloader ProxyDownloader
	//downloader.retryTimes = 10
	downloader.generator = generator
	return &downloader
}

func (this *ProxyDownloader) SetProxyGenerator(generator model.ProxyGenerator) {
	this.generator = generator
}

func (this *ProxyDownloader) Download(tag string, task model.Task) model.Result {
	var err error
	var resp *http.Response
	var result model.Result
	//var retry_times = 0
REDOWNLOAD:
	log.Println(tag, "Start Download: ", task.Url)
	var proxy model.Proxy
	if proxy = this.generator.GetProxy(); len(proxy.IP) == 0 {
		log.Println(tag, "You haven't set proxy.")
		return model.Result{Err: errors.New(ErrProxyNotSet)}
	} else {
		resp, err = dowloadWithProxy(task.Url, proxy)
	}

	if err != nil {
		log.Println(tag, "Download error occurred:", err.Error())
		this.generator.ChangeProxy(proxy)
		goto REDOWNLOAD
	}

	result.Identifier = task.Identifier
	result.Url = task.Url
	result.UserData = task.UserData
	result.Response, result.Err = model.GetResponse(resp)
	if result.Err != nil {
		log.Println(tag, "Getting Resp.Body error occurred:", result.Err.Error())
		goto REDOWNLOAD
	}
	log.Println(tag, "Download Success: ", task.Url)
	return result

}

func (this *ProxyDownloader) SetRetryTimes(times int) {
	//this.retryTimes = times
}
