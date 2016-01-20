package scheduler

import (
	"github.com/jinzhu/gorm"
	"github.com/nladuo/DLocker"
	"github.com/nladuo/go-webcrawler/model"
	"sync"
	"time"
)

const (
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

func newSqlScheduler(db *gorm.DB) *SqlScheduler {
	var scheduler SqlScheduler
	scheduler.db = db
	scheduler.tasks = make(chan model.Task, 300)
	scheduler.results = make(chan model.Result, 300)
	scheduler.addResultChan = make(chan byte, 300)
	scheduler.getResultChan = make(chan byte, 300)
	scheduler.addTaskChan = make(chan byte, 300)
	scheduler.getTaskChan = make(chan byte, 300)
	scheduler.flagGettingData = false
	createTable(db)
	return &scheduler
}

func NewDistributedSqlScheduler(db *gorm.DB, basePath, prefix string, timeout time.Duration) *SqlScheduler {
	scheduler := newSqlScheduler(db)
	scheduler.dLocker = DLocker.NewLocker(basePath, prefix, timeout)
	scheduler.isCluster = true
	go scheduler.manipulateDataLoop()
	return scheduler
}

func NewBasicSqlScheduler(db *gorm.DB) *SqlScheduler {
	scheduler := newSqlScheduler(db)
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
		case <-this.addTaskChan:
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
		case <-this.getTaskChan:
			if getTaskSize(this.db) < extract_count {
				// sleep 5 second
				for i := 0; i < 100; i++ {
					time.Sleep(50 * time.Millisecond)
					cleanChan(this.getTaskChan)
				}
			}
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
		case <-this.addResultChan:
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
		case <-this.getResultChan:
			if getResultSize(this.db) < extract_count {
				// sleep 5 second
				for i := 0; i < 100; i++ {
					time.Sleep(50 * time.Millisecond)
					cleanChan(this.getResultChan)
				}
			}
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
