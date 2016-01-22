package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	crawler "github.com/nladuo/go-webcrawler"
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"os"
	"strconv"
)

const (
	identifier string = "jikexueyuan"
)

func ParseCourse(res *model.Result, processor model.Processor) {
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
	if pageNum == 50 {
		os.Exit(0)
	}
	if pageNum == 1 {
		for i := 2; i < 52; i++ {
			pageNumStr := strconv.Itoa(i)
			task := model.Task{
				Identifier: identifier,
				Url:        "http://www.jikexueyuan.com/course/?pageNum=" + pageNumStr,
				UserData:   []byte(pageNumStr)}
			processor.AddTask(task)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "lack parameter")
		os.Exit(-1)
	}
	config, err := model.GetConfigFromPath(os.Args[1])
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open("postgres", "postgres://postgres:root@127.0.0.1/db_crawler?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	mCrawler := crawler.NewCrawler(&db, config)
	baseTask := model.Task{
		Identifier: identifier,
		Url:        "http://www.jikexueyuan.com/course/?pageNum=1",
		UserData:   []byte("1")}
	mCrawler.AddBaseTask(baseTask)
	parser := model.Parser{Identifier: identifier, Parse: ParseCourse}
	mCrawler.AddParser(parser)
	mCrawler.Run()
}
