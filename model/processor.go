package model

type Processor interface {
	AddTask(task Task)
}
