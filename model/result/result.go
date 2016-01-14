package result

import (
	"encoding/json"
	"github.com/nladuo/go-webcrawler/model/response"
)

type Result struct {
	Identifier string
	Err        error
	Response   response.HttpResponse
	UserData   []byte
}

func (this *Result) Serialize() (string, error) {
	resByte, err := json.Marshal(this)
	var resStr string
	if err == nil {
		resStr = string(resByte)
	}
	return resStr, err
}

func UnSerialize(data string) (Result, error) {
	var result Result
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}
