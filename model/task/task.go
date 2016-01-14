package task

import (
	"encoding/json"
)

type Task struct {
	Identifier string
	Url        string
	UserData   []byte
}

func (this *Task) Serialize() (string, error) {
	taskByte, err := json.Marshal(this)
	var taskStr string
	if err == nil {
		taskStr = string(taskByte)
	}
	return taskStr, err
}

func UnSerialize(data string) (Task, error) {
	var task Task
	err := json.Unmarshal([]byte(data), &task)
	return task, err
}
