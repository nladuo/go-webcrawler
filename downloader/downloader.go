package downloader

import (
	"github.com/nladuo/go-webcrawler/model/result"
	"github.com/nladuo/go-webcrawler/model/task"
	"io/ioutil"
	"net/http"
)

func Download(task task.Task) result.Result {
	var result result.Result
REDOWNLOAD:
	resp, err := http.Get(task.Url)
	if err != nil {
		goto REDOWNLOAD
	}
	//result.Response
	result.Identifier = task.Identifier
	result.UserData = task.UserData
	result.Response.Body, result.Err = ioutil.ReadAll(resp.Body)
	result.Response.Cookies = resp.Cookies()
	result.Response.Header = resp.Header
	return result
}
