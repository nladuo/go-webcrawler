//the local memory mode crawler
package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/nladuo/go-webcrawler"
	"github.com/nladuo/go-webcrawler/model"
)

const (
	threadNum                      = 3
	DOUBAN_MOVIE_TOP250_IDENTIFIER = "parse douban movie top250"
)

var (
	mCrawler *crawler.Crawler
)

func parse_movies(res model.Result, processor model.Processor) {
	doc, err := goquery.NewDocumentFromReader(res.Response.GetBodyReader())
	if err != nil {
		return
	}

	//get the repo name and its description
	doc.Find(".item").Each(func(i int, contentSelection *goquery.Selection) {
		movie_title := contentSelection.Find(".title").Text()
		fmt.Println(movie_title)
	})

	//the flag to check if the crawler finished.
	haveMorePages := false

	// add the next page task
	doc.Find(".next a").Each(func(i int, contentSelection *goquery.Selection) {
		nextPageHref, exists := contentSelection.Attr("href")
		if exists {
			nextPageHref = "https://movie.douban.com/top250" + nextPageHref
			processor.AddTask(model.Task{
				Url:        nextPageHref,
				Identifier: DOUBAN_MOVIE_TOP250_IDENTIFIER,
			})
			haveMorePages = true
		}
	})
	// if doesn't have more repos to crawl, stop the crawler
	if !haveMorePages {
		mCrawler.WaitForShutDown()
	}

}

func main() {
	//create a local memory mode crawler
	mCrawler = crawler.NewLocalMemCrawler(threadNum)

	//add initial task(s)
	firstPageTask := model.Task{
		Url:        "https://movie.douban.com/top250",
		Identifier: DOUBAN_MOVIE_TOP250_IDENTIFIER,
	}
	mCrawler.AddBaseTask(firstPageTask)

	// add parser(s) to handle the result(s) of task(s)
	mCrawler.AddParser(model.Parser{
		Identifier: DOUBAN_MOVIE_TOP250_IDENTIFIER,
		Parse:      parse_movies,
	})

	// start the crawler
	mCrawler.Run()
}
