package scheduler

import (
	"github.com/jinzhu/gorm"
	"github.com/nladuo/DLocker"
	"github.com/nladuo/go-webcrawler/model"
	"sync"
	"time"
)

var (
	//if the tasks or results length more than 200,
	//serialize the task and store it into sql database
	store_to_sql_count int = 200
	store_count        int = 100
	//if the tasks or results length less than 100,
	//get data from sql database
	extract_from_sql_count int = 100
	extract_count          int = 200
)

// the distributed scheduler
type SqlScheduler struct {
	db            *gorm.DB
	tasks         chan model.Task
	results       chan model.Result
	dLocker       *DLocker.Dlocker
	basicLocker   *sync.Mutex
	isCluster     bool
	getTaskChan   chan byte
	getResultChan chan byte
	addTaskChan   chan byte
	addResultChan chan byte
}

func newSqlScheduler(db *gorm.DB, bufferSize int) *SqlScheduler {
	var scheduler SqlScheduler
	scheduler.db = db
	store_to_sql_count = bufferSize / 2
	store_count = bufferSize
	extract_count = bufferSize / 2
	extract_count = bufferSize

	scheduler.tasks = make(chan model.Task, bufferSize+100)
	scheduler.results = make(chan model.Result, bufferSize+100)
	scheduler.addResultChan = make(chan byte, bufferSize+100)
	scheduler.getResultChan = make(chan byte, bufferSize+100)
	scheduler.addTaskChan = make(chan byte, bufferSize+100)
	scheduler.getTaskChan = make(chan byte, bufferSize+100)
	createTable(db)
	return &scheduler
}

func NewDistributedSqlScheduler(db *gorm.DB, bufferSize int, basePath, prefix string, timeout time.Duration) *SqlScheduler {
	scheduler := newSqlScheduler(db, bufferSize)
	scheduler.dLocker = DLocker.NewLocker(basePath, prefix, timeout)
	scheduler.isCluster = true
	go scheduler.manipulateDataLoop()
	return scheduler
}

func NewBasicSqlScheduler(db *gorm.DB, bufferSize int) *SqlScheduler {
	scheduler := newSqlScheduler(db, bufferSize)
	scheduler.basicLocker = &sync.Mutex{}
	scheduler.isCluster = false
	go scheduler.manipulateDataLoop()
	return scheduler
}

func (this *SqlScheduler) lock() {
	if this.isCluster {
		for !this.dLocker.Lock() {
		}
	} else {
		this.basicLocker.Lock()
	}
}

func (this *SqlScheduler) unLock() {
	if this.isCluster {
		this.dLocker.Unlock()
	} else {
		this.basicLocker.Unlock()
	}
}

func (this *SqlScheduler) manipulateDataLoop() {
	for {
		select {
		case <-this.addTaskChan: // add task does not need lock
			if len(this.tasks) > store_to_sql_count {
				this.lock()
				for i := 0; i < store_count; i++ {
					t := <-this.tasks
					taskStr, err := t.Serialize()
					if err != nil {
						continue
					}
					addTask(this.db, taskStr)
				}
				this.unLock()
			}
		case <-this.getTaskChan: // get task does need lock
			if len(this.tasks) < extract_from_sql_count {
				this.lock()
				tasks := getTasks(this.db, extract_count)
				for i := 0; i < len(tasks); i++ {
					t, err := model.UnSerializeTask(tasks[i].Data)
					if err != nil {
						continue
					}
					this.tasks <- t
				}
				this.unLock()
			}
		case <-this.addResultChan: // add result does not need lock
			if len(this.results) > store_to_sql_count {
				this.lock()
				for i := 0; i < store_count; i++ {
					r := <-this.results
					resultStr, err := r.Serialize()
					if err != nil {
						continue
					}
					addResult(this.db, resultStr)
				}
				this.unLock()
			}
		case <-this.getResultChan: // get result does need lock
			if len(this.results) < extract_from_sql_count {
				this.lock()
				results := getResults(this.db, extract_count)
				for i := 0; i < len(results); i++ {
					r, err := model.UnSerializeResult(results[i].Data)
					if err != nil {
						continue
					}
					this.results <- r
				}
				this.unLock()
			}
		}
	}
}

func (this *SqlScheduler) AddTask(task model.Task) {
	this.tasks <- task
	this.addTaskChan <- byte(1)
}

func (this *SqlScheduler) GetTask() model.Task {
	this.getTaskChan <- byte(1)
	return <-this.tasks
}

func (this *SqlScheduler) AddResult(result model.Result) {
	this.results <- result
	this.addResultChan <- byte(1)
}

func (this *SqlScheduler) GetResult() model.Result {
	this.getResultChan <- byte(1)
	return <-this.results
}
