package scheduler

import (
	"github.com/nladuo/go-webcrawler/model/result"
	"github.com/nladuo/go-webcrawler/model/task"
)

// the normal scheduler running in single pc
type Scheduler interface {
	AddTask(task task.Task)
	GetTask() task.Task
	AddResult(result result.Result)
	GetResult() result.Result
}
