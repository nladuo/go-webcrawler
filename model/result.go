package model

import (
	"encoding/json"
)

type Result struct {
	Identifier string
	Err        error
	Response   HttpResponse
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

func UnSerializeResult(data string) (Result, error) {
	var result Result
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}
