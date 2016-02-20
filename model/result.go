package model

import (
	"encoding/json"
)

type Result struct {
	Identifier string
	Err        error
	Url        string
	Response   *HttpResponse
	UserData   []byte
}

//if the user think the result has been verified by anti-spider.
// the user can change the proxy ip, and readd the initial task to the queue.
func (this *Result) GetInitialTask() *Task {
	task := Task{
		Identifier: this.Identifier,
		Url:        this.Url,
		UserData:   this.UserData,
	}
	return &task
}

func (this *Result) Serialize() (string, error) {
	resByte, err := json.Marshal(this)
	var resStr string
	if err == nil {
		resStr = string(resByte)
	}
	return resStr, err
}

func UnSerializeResult(data string) (Result, error) {
	var result Result
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}
