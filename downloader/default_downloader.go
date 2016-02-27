package downloader

import (
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"net/http"
)

const (
	DEFAULT_RETRY_TIMES = 10
)

type DefaultDownloader struct {
	retryTimes int
}

func NewDefaultDownloader() *DefaultDownloader {
	var downloader DefaultDownloader
	downloader.retryTimes = DEFAULT_RETRY_TIMES
	return &downloader
}

func (this *DefaultDownloader) Download(tag string, task model.Task) *model.Result {
	var err error
	var resp *http.Response
	var result model.Result
	var retry_times = 0
REDOWNLOAD:
	log.Println(tag, "Start Download: ", task.Url)
	if proxy := task.GetProxy(); len(proxy.IP) == 0 {
		resp, err = dowloadDirect(task.Url)
	} else {
		resp, err = dowloadWithProxy(task.Url, &proxy)
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
	if result.Err != nil {
		log.Println(tag, "Getting Resp.Body error occurred:", result.Err.Error())
		goto REDOWNLOAD
	}
	log.Println(tag, "Download Success: ", task.Url)
	return &result
}

func (this *DefaultDownloader) SetRetryTimes(times int) {
	this.retryTimes = times
}
