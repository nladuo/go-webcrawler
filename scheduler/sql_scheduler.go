package scheduler

import (
	"github.com/jinzhu/gorm"
	"github.com/nladuo/go-webcrawler/model"
	"github.com/nladuo/go-zk-lock"
	"log"
	"sync"
	"time"
)

//scheduler use sql database as task and result queue
type SqlScheduler struct {
	db          *gorm.DB
	tasks       chan model.Task
	dLocker     *DLocker.Dlocker
	basicLocker *sync.Mutex
	isCluster   bool
	getTaskChan chan byte
	addTaskChan chan byte
}

func newSqlScheduler(db *gorm.DB) *SqlScheduler {
	var scheduler SqlScheduler
	scheduler.db = db
	scheduler.tasks = make(chan model.Task, chan_buffer_size)
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
	}
}

func NewDistributedSqlScheduler(db *gorm.DB, basePath string, timeout time.Duration) *SqlScheduler {
	scheduler := newSqlScheduler(db)
	scheduler.dLocker = DLocker.NewLocker(basePath, timeout)
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

func (this *SqlScheduler) GetTaskSize() int {
	return len(this.tasks)
}

// the serialization for model.Task in sql database
type Task struct {
	ID   uint   `sql:"AUTO_INCREMENT"`
	Data string `sql:"size:max"`
}

func createTable(db *gorm.DB) {
	if !db.HasTable(&Task{}) {
		db.CreateTable(&Task{})
	}
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
			log.Println("scheduler.getTasks---->rollback")
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
