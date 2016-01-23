package scheduler

import (
	"github.com/nladuo/go-webcrawler/model"
)

// the interface for manipulation of tasks and results
type Scheduler interface {
	AddTask(task model.Task)
	GetTask() model.Task
	AddResult(result model.Result)
	GetResult() model.Result
}
