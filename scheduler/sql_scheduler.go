package scheduler

import (
	"github.com/jinzhu/gorm"
	"github.com/nladuo/DLocker"
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"sync"
	"time"
)

//scheduler use sql database as task and result queue
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
	scheduler.tasks = make(chan model.Task, chan_buffer_size)
	scheduler.results = make(chan model.Result, chan_buffer_size)
	scheduler.addResultChan = make(chan byte, chan_buffer_size)
	scheduler.getResultChan = make(chan byte, chan_buffer_size)
	scheduler.addTaskChan = make(chan byte, chan_buffer_size)
	scheduler.getTaskChan = make(chan byte, chan_buffer_size)
	createTable(db)
	go scheduler.logTaskAndResultNum()
	return &scheduler
}

func (this *SqlScheduler) logTaskAndResultNum() {
	for {
		time.Sleep(3 * time.Minute)
		log.Println("task num:", len(this.tasks))
		log.Println("result num:", len(this.results))
	}
}

func NewDistributedSqlScheduler(db *gorm.DB, basePath, prefix string, timeout time.Duration) *SqlScheduler {
	scheduler := newSqlScheduler(db)
	scheduler.dLocker = DLocker.NewLocker(basePath, prefix, timeout)
	scheduler.isCluster = true
	go scheduler.manipulateDataLoop()
	return scheduler
}

func NewLocalSqlScheduler(db *gorm.DB) *SqlScheduler {
	scheduler := newSqlScheduler(db)
	scheduler.basicLocker = &sync.Mutex{}
	scheduler.isCluster = false
	go scheduler.manipulateDataLoop()
	return scheduler
}

func (this *SqlScheduler) lock() {
	if this.isCluster {
		this.dLocker.Lock()
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

//serialize data(task or result) and store into sql db.
//  or unserialize data from sql db.
func (this *SqlScheduler) manipulateDataLoop() {
	for {
		select {
		case <-this.addTaskChan: // add task does not need lock
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
		case <-this.getTaskChan: // get task does need lock
			this.lock()
			if len(this.tasks) < extract_count {
				tasks := getTasks(this.db, extract_count)
				for i := 0; i < len(tasks); i++ {
					t, err := model.UnSerializeTask(tasks[i].Data)
					if err != nil {
						continue
					}
					this.tasks <- t
				}
			}
			this.unLock()
		case <-this.addResultChan: // add result does not need lock
			if len(this.results) > store_to_sql_count {
				for i := 0; i < store_count; i++ {
					r := <-this.results
					resultStr, err := r.Serialize()
					if err != nil {
						continue
					}
					addResult(this.db, resultStr)
				}
			}
		case <-this.getResultChan: // get result does need lock
			this.lock()
			if len(this.results) < extract_from_sql_count {
				results := getResults(this.db, extract_count)
				for i := 0; i < len(results); i++ {
					r, err := model.UnSerializeResult(results[i].Data)
					if err != nil {
						continue
					}
					this.results <- r
				}
			}
			this.unLock()
		}
	}
}

func (this *SqlScheduler) AddTask(task model.Task) {
	if len(this.tasks) > store_to_sql_count {
		this.addTaskChan <- byte(1)
	}
	this.tasks <- task
}

func (this *SqlScheduler) GetTask() model.Task {
	if len(this.tasks) < extract_from_sql_count {
		this.getTaskChan <- byte(1)
	}
	return <-this.tasks
}

func (this *SqlScheduler) AddResult(result model.Result) {
	if len(this.results) > store_to_sql_count {
		this.addResultChan <- byte(1)
	}
	this.results <- result
}

func (this *SqlScheduler) GetResult() model.Result {
	if len(this.results) < extract_from_sql_count {
		this.getResultChan <- byte(1)
	}
	return <-this.results
}

// the serialization for model.Task in sql database
type Task struct {
	ID   uint   `sql:"AUTO_INCREMENT"`
	Data string `sql:"size:max"`
}

// the serialization for model.Result in sql database
type Result struct {
	ID   uint   `sql:"AUTO_INCREMENT"`
	Data string `sql:"size:max"`
}

func createTable(db *gorm.DB) {
	if !db.HasTable(&Task{}) {
		db.CreateTable(&Task{})
	}
	if !db.HasTable(&Result{}) {
		db.CreateTable(&Result{})
	}
}

func addResult(db *gorm.DB, data string) {
	db.Create(&Result{Data: data})
}

func getResults(db *gorm.DB, limit int) []Result {
	results := []Result{}
	t := db.Begin()
	t.Limit(limit).Find(&results)
	for i := 0; i < len(results); i++ {
		rowsAffected := t.Delete(&results[i]).RowsAffected
		if rowsAffected == 0 {
			t.Rollback()
			log.Println("getResults---->rollback")
			return []Result{}
		}
	}
	t.Commit()
	return results
}

func getResultSize(db *gorm.DB) int {
	count := 0
	db.Model(&Result{}).Count(&count)
	return count
}

func addTask(db *gorm.DB, data string) {
	db.Create(&Task{Data: data})
}

func getTasks(db *gorm.DB, limit int) []Task {
	tasks := []Task{}
	t := db.Begin()
	//get tasks form db
	t.Limit(limit).Find(&tasks)
	// delete the tasks
	for i := 0; i < len(tasks); i++ {
		rowsAffected := t.Delete(&tasks[i]).RowsAffected
		if rowsAffected == 0 {
			t.Rollback()
			log.Println("getTasks---->rollback")
			return []Task{}
		}
	}
	t.Commit()
	return tasks
}

func getTaskSize(db *gorm.DB) int {
	count := 0
	db.Model(&Task{}).Count(&count)
	return count
}
