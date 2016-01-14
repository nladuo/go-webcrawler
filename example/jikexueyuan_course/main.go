package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/nladuo/go-webcrawler/crawler"
	"github.com/nladuo/go-webcrawler/model/config"
	"os"
)

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
	crawler := crawler.NewCrawler(&db, config)
	crawler.AddBaseTask(task)
	crawler.AddParser(parser)
}
