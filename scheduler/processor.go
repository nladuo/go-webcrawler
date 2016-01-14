package scheduler

import (
	"github.com/nladuo/go-webcrawler/model/task"
)

type Processor interface {
	AddTask(task task.Task)
}
