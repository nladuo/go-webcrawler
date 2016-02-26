package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	crawler "github.com/nladuo/go-webcrawler"
	"github.com/nladuo/go-webcrawler/model"
	"log"
	"os"
	"strconv"
)

const (
	PARSE_COURSE_URL    string = "parse_course_url"
	PARSE_COURSE_DETAIL string = "parse_course_detail"
	threadNum           int    = 3
)

var mCrawler *crawler.Crawler

func ParseCourseDetail(res *model.Result, processor model.Processor) {

}

func ParseCourseUrl(res *model.Result, processor model.Processor) {
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
				Identifier: PARSE_COURSE_URL,
				Url:        "http://www.jikexueyuan.com/course/?pageNum=" + pageNumStr,
				UserData:   []byte(pageNumStr)}
			processor.AddTask(task)
		}
	}
}

func main() {
	mCrawler = crawler.NewLocalMemCrawler(threadNum)
	baseTask := model.Task{
		Identifier: PARSE_COURSE_URL,
		Url:        "http://www.jikexueyuan.com/course/?pageNum=1",
		UserData:   []byte("1")}
	mCrawler.AddBaseTask(baseTask)
	parser := model.Parser{Identifier: PARSE_COURSE_URL, Parse: ParseCourseUrl}
	mCrawler.AddParser(parser)
	mCrawler.Run()
}
