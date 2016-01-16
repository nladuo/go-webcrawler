package scheduler

import (
	"github.com/nladuo/go-webcrawler/model"
)

// the normal scheduler running in single pc
type Scheduler interface {
	AddTask(task model.Task)
	GetTask() model.Task
	AddResult(result model.Result)
	GetResult() model.Result
}
