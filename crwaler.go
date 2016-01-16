package crawler

import (
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
	isCluster bool
}

func NewCrawler(db *gorm.DB, config *model.Config) *Crawler {
	var crawler Crawler
	crawler.isCluster = config.IsCluster
	crawler.isMaster = config.IsMaster
	crawler.threadNum = config.ThreadNum
	appName = config.AppName
	zkHosts = config.ZkHosts
	lockersPath = "/" + appName + "/" + lockerDir
	lockerTimeout = time.Duration(config.LockerTimeout) * time.Second
	zkTimeOut = time.Duration(config.ZkTimeOut) * time.Second
	if crawler.isCluster {
		DLocker.EstablishZkConn(zkHosts, zkTimeOut)
		DLocker.CreatePath("/" + appName)
		sqlScheduler := scheduler.NewDistributedSqlScheduler(db, lockersPath, prefix, lockerTimeout)
		crawler.scheduler = sqlScheduler
		crawler.processor = sqlScheduler
	} else {
		sqlScheduler := scheduler.NewBasicSqlScheduler(db)
		crawler.scheduler = sqlScheduler
		crawler.processor = sqlScheduler
	}
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
		go func() {
			for {
				task := this.scheduler.GetTask()
				result := downloader.Download(task)
				this.scheduler.AddResult(result)
			}
		}()
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
