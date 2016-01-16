package downloader

import (
	"github.com/nladuo/go-webcrawler/model"
	"io/ioutil"
	"log"
	"net/http"
)

func Download(task model.Task) model.Result {
	var result model.Result
REDOWNLOAD:
	resp, err := http.Get(task.Url)
	if err != nil {
		goto REDOWNLOAD
	}
	log.Println("Download: ", task.Url, task.Identifier)
	//result.Response
	result.Identifier = task.Identifier
	result.UserData = task.UserData
	result.Response.Body, result.Err = ioutil.ReadAll(resp.Body)
	result.Response.Cookies = resp.Cookies()
	result.Response.Header = resp.Header
	return result
}
