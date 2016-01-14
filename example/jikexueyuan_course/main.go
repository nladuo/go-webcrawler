package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/nladuo/go-webcrawler/crawler"
	"github.com/nladuo/go-webcrawler/model/config"
	"github.com/nladuo/go-webcrawler/model/result"
	"github.com/nladuo/go-webcrawler/model/task"
	"github.com/nladuo/go-webcrawler/scheduler"
	"log"
	"os"
	"strconv"
)

const (
	identifier string = "jikexueyuan"
)

func ParseCourse(res *result.Result, processor scheduler.Processor) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Response.Body))
	if err != nil {
		return
	}

	doc.Find(".lessonimg-box a").Each(func(i int, contentSelection *goquery.Selection) {
		title, _ := contentSelection.Find("img").Attr("title")
		fmt.Println(title)
	})
	pageNum, _ := strconv.Atoi(string(res.UserData))
	log.Println("page num :", pageNum)
	if pageNum > 50 {
		os.Exit(0)
	}
	pageNumStr := strconv.Itoa(pageNum + 1)
	task := task.Task{
		Identifier: identifier,
		Url:        "http://www.jikexueyuan.com/course/?pageNum=" + pageNumStr,
		UserData:   []byte(pageNumStr)}
	processor.AddTask(task)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "lack parameter")
		os.Exit(-1)
	}
	config, err := config.GetConfigFromPath(os.Args[1])
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open("postgres", "postgres://postgres:root@127.0.0.1/db_crawler?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	mCrawler := crawler.NewCrawler(&db, config)
	baseTask := task.Task{
		Identifier: identifier,
		Url:        "http://www.jikexueyuan.com/course/?pageNum=1",
		UserData:   []byte("1")}
	mCrawler.AddBaseTask(baseTask)
	parser := crawler.Parser{Identifier: identifier, Parse: ParseCourse}
	mCrawler.AddParser(parser)
	mCrawler.Run()
}
