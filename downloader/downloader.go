package downloader

import (
	"github.com/nladuo/go-webcrawler/model"
	"io/ioutil"
	"log"
	"net/http"
)

func Download(tag string, task model.Task) model.Result {
	log.Println(tag, "Start Download: ", task.Url)
	var result model.Result
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
	log.Println(tag, "Download Success: ", task.Url)
	return result
}
