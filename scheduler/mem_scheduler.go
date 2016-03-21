package scheduler

import (
	"container/list"
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"sync"
	"time"
)

//scheduler use memory as task and result queue
type LocalMemScheduler struct {
	locker      *sync.Mutex
	taskList    *list.List
	tasks       chan model.Task
	getTaskChan chan byte
	addTaskChan chan byte
}

func NewLocalMemScheduler() *LocalMemScheduler {
	var scheduler LocalMemScheduler
	scheduler.locker = &sync.Mutex{}
	scheduler.taskList = list.New()
	scheduler.tasks = make(chan model.Task, chan_buffer_size)
	scheduler.addTaskChan = make(chan byte, chan_buffer_size)
	scheduler.getTaskChan = make(chan byte, chan_buffer_size)
	go scheduler.logTaskAndResultNum()
	go scheduler.manipulateDataLoop()
	return &scheduler
}

func (this *LocalMemScheduler) logTaskAndResultNum() {
	for {
		time.Sleep(3 * time.Minute)
		log.Println("task num:", len(this.tasks))
	}
}

func (this *LocalMemScheduler) manipulateDataLoop() {
	for {
		select {
		case <-this.addTaskChan:
			this.locker.Lock()
			if len(this.tasks) > store_to_sql_count {
				for i := 0; i < store_count; i++ {
					t := <-this.tasks
					this.taskList.PushBack(t)
				}
			}
			this.locker.Unlock()
		case <-this.getTaskChan:
			this.locker.Lock()
			if len(this.tasks) < extract_count {
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
			}
			this.locker.Unlock()
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

func (this *LocalMemScheduler) GetTaskSize() int {
	return len(this.tasks)
}
