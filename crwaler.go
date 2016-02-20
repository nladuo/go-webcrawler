package crawler

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nladuo/DLocker"
	"github.com/nladuo/go-webcrawler/downloader"
	"github.com/nladuo/go-webcrawler/model"
	"github.com/nladuo/go-webcrawler/scheduler"
	"log"
	"time"
)

// arguments for zookeeper
const (
	prefix    string = "lock-"
	lockerDir string = "locker"
)

// arguments for zookeeper
var (
	appName       string
	zkHosts       []string
	lockersPath   string
	lockerTimeout time.Duration
	zkTimeOut     time.Duration
)

type Crawler struct {
	threadNum  int
	parsers    []*model.Parser
	scheduler  scheduler.Scheduler
	processor  model.Processor
	downloader downloader.Downloader
	isMaster   bool //only the master crawler excute the Crawler.AddBaseTask
}

//used for distributed mode,need zookeeper
// and a sql database to store the internal data.
// Make sure you sql database can be accessed by all the server
func NewDistributedSqlCrawler(db *gorm.DB, config *model.DistributedConfig) *Crawler {
	var crawler Crawler
	crawler.isMaster = config.IsMaster
	crawler.threadNum = config.ThreadNum

	//setup the configuration for zookeeper
	appName = config.AppName
	zkHosts = config.ZkHosts
	lockersPath = "/" + appName + "/" + lockerDir
	lockerTimeout = time.Duration(config.LockerTimeout) * time.Second
	zkTimeOut = time.Duration(config.ZkTimeOut) * time.Second

	//setup the distributed locker
	DLocker.EstablishZkConn(zkHosts, zkTimeOut)
	DLocker.CreatePath("/" + appName)

	//initialize the scheduler
	sqlScheduler := scheduler.NewDistributedSqlScheduler(db, lockersPath, prefix, lockerTimeout)
	crawler.scheduler = sqlScheduler
	crawler.processor = sqlScheduler

	//set the downloader to nil
	crawler.downloader = nil
	return &crawler
}

//local mode, need a sql database to store the internal data
// and to spare the memory use.
func NewLocalSqlCrawler(db *gorm.DB, threadNum int) *Crawler {
	var crawler Crawler
	//setup the configuration
	crawler.isMaster = true
	crawler.threadNum = threadNum

	//initialize the scheduler
	sqlScheduler := scheduler.NewLocalSqlScheduler(db)
	crawler.scheduler = sqlScheduler
	crawler.processor = sqlScheduler

	//set the downloader to nil
	crawler.downloader = nil
	return &crawler
}

//local mode, store the internal data into a queue,
// suitable for simple application.
func NewLocalMemCrawler(threadNum int) *Crawler {
	var crawler Crawler
	crawler.isMaster = true
	crawler.threadNum = threadNum

	//initialize the scheduler
	memScheduler := scheduler.NewLocalMemScheduler()
	crawler.scheduler = memScheduler
	crawler.processor = memScheduler

	//set the downloader to nil
	crawler.downloader = nil
	return &crawler
}

//only the master crawler excute the Crawler.AddBaseTask
// So, if you are under the Distributed Mode,
// you can just change the config.json and make your crawler work distributedly.
func (this *Crawler) AddBaseTask(task model.Task) {
	if this.isMaster {
		this.scheduler.AddTask(task)
	}
}

func (this *Crawler) AddParser(parser model.Parser) {
	this.parsers = append(this.parsers, &parser)
}

func (this *Crawler) SetProxyGenerator(generater model.ProxyGenerator) {
	this.downloader = downloader.NewProxyDownloader(generater)
}

func (this *Crawler) Run() {
	if this.downloader == nil {
		this.downloader = downloader.NewDefaultDownloader()
	}
	// netWork Handle goroutine
	for i := 0; i < this.threadNum; i++ {
		go func(num int) {
			tag := fmt.Sprintf("[goroutine %d]", num)
			for {
				task := this.scheduler.GetTask()
				result := this.downloader.Download(tag, task)
				if result == nil {
					continue
				}
				this.scheduler.AddResult(*result)
			}
		}(i + 1)
	}

	//parser the result
	for {
		result := this.scheduler.GetResult()
		log.Println("Get task, Identifier: " + result.Identifier)
		for i := 0; i < len(this.parsers); i++ {
			if this.parsers[i].Identifier == result.Identifier {
				this.parsers[i].Parse(&result, this.processor)
				break
			}
		}
	}
}
