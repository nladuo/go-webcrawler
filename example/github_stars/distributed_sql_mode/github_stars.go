//the distributed sql mode crawler
package main

import (
	"fmt"
	_ "github.com/Go-SQL-Driver/MySQL"
	"github.com/PuerkitoBio/goquery"
	"github.com/jinzhu/gorm"
	"github.com/nladuo/go-webcrawler"
	"github.com/nladuo/go-webcrawler/model"
	"os"
	"strings"
)

const (
	GITHUB_STARS_IDENTIFIER = "parse github stars"
)

var (
	mCrawler *crawler.Crawler
)

// When a task with identifier named "GITHUB_STARS_IDENTIFIER" have been downloaded, this function will be called.
func parse_repos(res model.Result, processor model.Processor) {
	doc, err := goquery.NewDocumentFromReader(res.Response.GetBodyReader())
	if err != nil {
		return
	}

	//get the repo name and its description
	doc.Find("li[class='repo-list-item public source']").Each(func(i int, contentSelection *goquery.Selection) {
		repo_name, _ := contentSelection.Find(".repo-list-name a").Attr("href")
		repo_description := contentSelection.Find(".repo-list-description").Text()
		repo_description = strings.Trim(repo_description, "\n")
		repo_description = strings.Trim(repo_description, " ")
		fmt.Println(repo_name, ":\n", repo_description)
	})

	//the flag to check if the crawler finished.
	haveMorePages := false

	// add the next page task
	doc.Find(".paginate-container .pagination a").Each(func(i int, contentSelection *goquery.Selection) {
		if contentSelection.Text() == "Next" {
			nextPageHref, exists := contentSelection.Attr("href")
			if exists {
				processor.AddTask(model.Task{
					Url:        nextPageHref,
					Identifier: GITHUB_STARS_IDENTIFIER,
				})
				haveMorePages = true
			}
		}
	})
	// if doesn't have more repos to crawl, stop the crawler
	if !haveMorePages {
		mCrawler.WaitForShutDown()
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "lack parameter")
		os.Exit(-1)
	}
	//read the config.json file
	config, err := model.GetDistributedConfigFromPath(os.Args[1])
	if err != nil {
		panic(err)
	}

	//create a db named test with charset=utf8
	db, err := gorm.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True")
	if err != nil {
		panic(err)
	}

	//create a distributed sql mode crawler
	mCrawler = crawler.NewDistributedSqlCrawler(db, config)

	//add initial task(s)
	firstPageTask := model.Task{
		Url:        "https://github.com/stars/nladuo",
		Identifier: GITHUB_STARS_IDENTIFIER,
	}
	mCrawler.AddBaseTask(firstPageTask)

	// add parser(s) to handle the result(s) of task(s)
	mCrawler.AddParser(model.Parser{
		Identifier: GITHUB_STARS_IDENTIFIER,
		Parse:      parse_repos,
	})

	// start the crawler
	mCrawler.Run()
}
