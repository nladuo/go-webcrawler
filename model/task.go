package model

import (
	"encoding/json"
)

type Task struct {
	Identifier string
	Url        string
	UserData   []byte
	proxy      Proxy
}

func (this *Task) SetProxy(proxy Proxy) {
	this.proxy = proxy
}

func (this *Task) GetProxy() Proxy {
	return this.proxy
}

func (this *Task) Serialize() (string, error) {
	taskByte, err := json.Marshal(this)
	var taskStr string
	if err == nil {
		taskStr = string(taskByte)
	}
	return taskStr, err
}

func UnSerializeTask(data string) (Task, error) {
	var task Task
	err := json.Unmarshal([]byte(data), &task)
	return task, err
}
