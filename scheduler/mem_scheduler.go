package scheduler

import (
	"container/list"
	"github.com/nladuo/go-webcrawler/model"
	"sync"
)

//scheduler use memory as task and result queue
type LocalMemScheduler struct {
	locker        *sync.Mutex
	taskList      *list.List
	resultList    *list.List
	tasks         chan model.Task
	results       chan model.Result
	getTaskChan   chan byte
	getResultChan chan byte
	addTaskChan   chan byte
	addResultChan chan byte
}

func NewLocalMemScheduler() *LocalMemScheduler {
	var scheduler LocalMemScheduler
	scheduler.locker = &sync.Mutex{}
	scheduler.taskList = list.New()
	scheduler.resultList = list.New()
	scheduler.tasks = make(chan model.Task, chan_buffer_size)
	scheduler.results = make(chan model.Result, chan_buffer_size)
	scheduler.addResultChan = make(chan byte, chan_buffer_size)
	scheduler.getResultChan = make(chan byte, chan_buffer_size)
	scheduler.addTaskChan = make(chan byte, chan_buffer_size)
	scheduler.getTaskChan = make(chan byte, chan_buffer_size)
	go scheduler.manipulateDataLoop()
	return &scheduler
}

func (this *LocalMemScheduler) manipulateDataLoop() {
	for {
		select {
		case <-this.addTaskChan:
			if len(this.tasks) > store_to_sql_count {
				this.locker.Lock()
				for i := 0; i < store_count; i++ {
					t := <-this.tasks
					this.taskList.PushBack(t)
				}
				this.locker.Unlock()
			}
		case <-this.getTaskChan:
			if len(this.tasks) < extract_count {
				this.locker.Lock()
				popCount := extract_count
				if this.taskList.Len() < extract_count {
					popCount = this.taskList.Len()
				}
				for i := 0; i < popCount; i++ {
					e := this.taskList.Front() //get the first task
					t := e.Value.(model.Task)
					this.taskList.Remove(e) // delete the first task
					this.tasks <- t
				}
				this.locker.Unlock()
			}
		case <-this.addResultChan:
			if len(this.results) > store_to_sql_count {
				this.locker.Lock()
				for i := 0; i < store_count; i++ {
					r := <-this.results
					this.resultList.PushBack(r)
				}
				this.locker.Unlock()
			}
		case <-this.getResultChan:
			if len(this.results) < extract_from_sql_count {
				this.locker.Lock()
				popCount := extract_count
				if this.taskList.Len() < extract_count {
					popCount = this.taskList.Len()
				}
				for i := 0; i < popCount; i++ {
					e := this.resultList.Front() //get the first task
					r := e.Value.(model.Result)
					this.resultList.Remove(e) // delete the first task
					this.results <- r
				}
				this.locker.Unlock()
			}
		}
	}
}

func (this *LocalMemScheduler) AddTask(task model.Task) {
	if len(this.tasks) > store_to_sql_count {
		this.addTaskChan <- byte(1)
	}
	this.tasks <- task
}

func (this *LocalMemScheduler) GetTask() model.Task {
	if len(this.tasks) < extract_from_sql_count {
		this.getTaskChan <- byte(1)
	}
	return <-this.tasks
}

func (this *LocalMemScheduler) AddResult(result model.Result) {
	if len(this.results) > store_to_sql_count {
		this.addResultChan <- byte(1)
	}
	this.results <- result
}

func (this *LocalMemScheduler) GetResult() model.Result {
	if len(this.results) < extract_from_sql_count {
		this.getResultChan <- byte(1)
	}
	return <-this.results
}
