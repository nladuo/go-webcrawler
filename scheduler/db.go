package scheduler

import (
	"github.com/jinzhu/gorm"
	"log"
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
	t := db.Begin()
	t.Limit(limit).Find(&results)
	for i := 0; i < len(results); i++ {
		rowsAffected := t.Delete(&results[i]).RowsAffected
		log.Println("rowsAffected--->", rowsAffected)
		if rowsAffected == 0 {
			t.Rollback()
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
		log.Println("rowsAffected--->", rowsAffected)
		if rowsAffected == 0 {
			t.Rollback()
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
