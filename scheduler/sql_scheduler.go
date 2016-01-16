package scheduler

import (
	"github.com/jinzhu/gorm"
	"github.com/nladuo/DLocker"
	"github.com/nladuo/go-webcrawler/model/result"
	"github.com/nladuo/go-webcrawler/model/task"
	"sync"
	"time"
)

const (
	//if the tasks or results length more than 120,
	//serialize the task and store it into sql database
	store_to_sql_count int = 120
	store_count        int = 100
	//if the tasks or results length less than 20,
	//get data from sql database
	extract_from_sql_count int = 20
	extract_count          int = 100
)

// the distributed scheduler
type SqlScheduler struct {
	db          *gorm.DB
	tasks       chan task.Task
	results     chan result.Result
	dLocker     *DLocker.Dlocker
	basicLocker *sync.Mutex
	isCluster   bool
}

func newSqlScheduler(db *gorm.DB) *SqlScheduler {
	var scheduler SqlScheduler
	scheduler.db = db
	scheduler.tasks = make(chan task.Task, 200)
	scheduler.results = make(chan result.Result, 200)
	createTable(db)
	return &scheduler
}

func NewDistributedSqlScheduler(db *gorm.DB, basePath, prefix string, timeout time.Duration) *SqlScheduler {
	scheduler := newSqlScheduler(db)
	scheduler.dLocker = DLocker.NewLocker(basePath, prefix, timeout)
	scheduler.isCluster = true
	return scheduler
}

func NewBasicSqlScheduler(db *gorm.DB) *SqlScheduler {
	scheduler := newSqlScheduler(db)
	scheduler.basicLocker = &sync.Mutex{}
	scheduler.isCluster = false
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

// add task into scheduler,
// if the tasks size exceeds 120,
// some tasks will store into database
func (this *SqlScheduler) AddTask(task task.Task) {
	this.tasks <- task
	go func() {
		this.lock()
		if len(this.tasks) > store_to_sql_count {
			for i := 0; i < store_count; i++ {
				t := <-this.tasks
				taskStr, err := t.Serialize()
				if err != nil {
					continue
				}
				addTask(this.db, taskStr)
			}
		}
		this.unLock()
	}()
}

// get task into scheduler,
// if the tasks size less than 20,
// it will extract some tasks from database
func (this *SqlScheduler) GetTask() task.Task {
	go func() {
		this.lock()
		if len(this.tasks) < extract_from_sql_count {
			tasks := getTasks(this.db, extract_count)
			for i := 0; i < len(tasks); i++ {
				t, err := task.UnSerialize(tasks[i].Data)
				if err != nil {
					continue
				}
				this.tasks <- t
			}

		}
		this.unLock()
	}()
	return <-this.tasks
}

func (this *SqlScheduler) AddResult(result result.Result) {
	this.results <- result
	go func() {
		this.lock()
		if len(this.tasks) > store_to_sql_count {
			for i := 0; i < store_count; i++ {
				r := <-this.results
				resultStr, err := r.Serialize()
				if err != nil {
					continue
				}
				addResult(this.db, resultStr)
			}
		}
		this.unLock()
	}()
}

func (this *SqlScheduler) GetResult() result.Result {
	go func() {
		this.lock()
		if len(this.tasks) < extract_from_sql_count {
			results := getResults(this.db, extract_count)
			for i := 0; i < len(results); i++ {
				r, err := result.UnSerialize(results[i].Data)
				if err != nil {
					continue
				}
				this.results <- r
			}
		}
		this.unLock()
	}()
	return <-this.results
}
