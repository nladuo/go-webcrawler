package crawler

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nladuo/DLocker"
	"github.com/nladuo/go-webcrawler/downloader"
	"github.com/nladuo/go-webcrawler/model"
	"github.com/nladuo/go-webcrawler/scheduler"
	"time"
)

const (
	prefix    string = "lock-"
	lockerDir string = "locker"
)

var (
	appName       string
	zkHosts       []string
	lockersPath   string
	lockerTimeout time.Duration
	zkTimeOut     time.Duration
)

type Crawler struct {
	threadNum int
	parsers   []*model.Parser
	scheduler scheduler.Scheduler
	processor model.Processor
	isMaster  bool
}

func NewDistributedSqlCrawler(db *gorm.DB, config *model.DistributedConfig) *Crawler {
	var crawler Crawler
	crawler.isMaster = config.IsMaster
	crawler.threadNum = config.ThreadNum
	appName = config.AppName
	zkHosts = config.ZkHosts
	lockersPath = "/" + appName + "/" + lockerDir
	lockerTimeout = time.Duration(config.LockerTimeout) * time.Second
	zkTimeOut = time.Duration(config.ZkTimeOut) * time.Second
	DLocker.EstablishZkConn(zkHosts, zkTimeOut)
	DLocker.CreatePath("/" + appName)
	sqlScheduler := scheduler.NewDistributedSqlScheduler(db, lockersPath, prefix, lockerTimeout)
	crawler.scheduler = sqlScheduler
	crawler.processor = sqlScheduler
	return &crawler
}

func NewLocalSqlCrawler(db *gorm.DB, threadNum int) *Crawler {
	var crawler Crawler
	crawler.isMaster = true
	crawler.threadNum = threadNum
	sqlScheduler := scheduler.NewLocalSqlScheduler(db)
	crawler.scheduler = sqlScheduler
	crawler.processor = sqlScheduler
	return &crawler
}

func NewLocalMemCrawler(threadNum int) *Crawler {
	var crawler Crawler
	crawler.isMaster = true
	crawler.threadNum = threadNum
	memScheduler := scheduler.NewLocalMemScheduler()
	crawler.scheduler = memScheduler
	crawler.processor = memScheduler
	return &crawler
}

func (this *Crawler) AddBaseTask(task model.Task) {
	if this.isMaster {
		this.scheduler.AddTask(task)
	}
}

func (this *Crawler) AddParser(parser model.Parser) {
	this.parsers = append(this.parsers, &parser)
}

func (this *Crawler) Run() {

	for i := 0; i < this.threadNum; i++ {
		go func(num int) {
			tag := fmt.Sprintf("[goroutine %d]", num)
			for {
				task := this.scheduler.GetTask()
				result := downloader.Download(tag, task)
				this.scheduler.AddResult(result)
			}
		}(i + 1)
	}

	for {
		result := this.scheduler.GetResult()
		for i := 0; i < len(this.parsers); i++ {
			if this.parsers[i].Identifier == result.Identifier {
				this.parsers[i].Parse(&result, this.processor)
				break
			}
		}
	}
}
