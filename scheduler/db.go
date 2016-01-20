package scheduler

import (
	"github.com/jinzhu/gorm"
)

type Task struct {
	ID   uint   `sql:"AUTO_INCREMENT"`
	Data string `sql:"size:max"`
}

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
	db.Limit(limit).Find(&results)
	for i := 0; i < len(results); i++ {
		db.Delete(&results[i])
	}
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
	//get tasks form db
	db.Limit(limit).Find(&tasks)
	// delete the tasks
	for i := 0; i < len(tasks); i++ {
		db.Delete(&tasks[i])
	}
	return tasks
}

func getTaskSize(db *gorm.DB) int {
	count := 0
	db.Model(&Task{}).Count(&count)
	return count
}
