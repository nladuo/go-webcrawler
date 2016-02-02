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
	downloadThreadNum int
	parseThreadNum    int //defaut is only one goroutine to parse result
	parsers           []*model.Parser
	scheduler         scheduler.Scheduler
	processor         model.Processor
	isMaster          bool
	end               chan byte // unbufferred channal
}

func NewDistributedSqlCrawler(db *gorm.DB, config *model.DistributedConfig) *Crawler {
	var crawler Crawler
	crawler.isMaster = config.IsMaster
	crawler.downloadThreadNum = config.ThreadNum
	crawler.parseThreadNum = 1
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
	crawler.end = make(chan byte)
	return &crawler
}

func NewLocalSqlCrawler(db *gorm.DB, downloadThreadNum int) *Crawler {
	var crawler Crawler
	crawler.isMaster = true
	crawler.downloadThreadNum = downloadThreadNum
	crawler.parseThreadNum = 1
	sqlScheduler := scheduler.NewLocalSqlScheduler(db)
	crawler.scheduler = sqlScheduler
	crawler.processor = sqlScheduler
	crawler.end = make(chan byte)
	return &crawler
}

func NewLocalMemCrawler(downloadThreadNum int) *Crawler {
	var crawler Crawler
	crawler.isMaster = true
	crawler.downloadThreadNum = downloadThreadNum
	crawler.parseThreadNum = 1
	memScheduler := scheduler.NewLocalMemScheduler()
	crawler.scheduler = memScheduler
	crawler.processor = memScheduler
	crawler.end = make(chan byte)
	return &crawler
}

// Stop the program
func (this *Crawler) ShutDown() {
	this.end <- byte(1)
}

func (this *Crawler) SetParseThreadNum(threadNum int) {
	this.parseThreadNum = threadNum
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

	// netWork Handle goroutine
	for i := 0; i < this.downloadThreadNum; i++ {
		go func(num int) {
			tag := fmt.Sprintf("[goroutine %d]", num)
			for {
				task := this.scheduler.GetTask()
				result := downloader.Download(tag, task)
				if result == nil {
					continue
				}
				this.scheduler.AddResult(*result)
			}
		}(i + 1)
	}

	//parser goroutine
	for i := 0; i < this.parseThreadNum; i++ {

		go func() {
			for {
				result := this.scheduler.GetResult()
				for i := 0; i < len(this.parsers); i++ {
					if this.parsers[i].Identifier == result.Identifier {
						this.parsers[i].Parse(&result, this.processor)
						break
					}
				}
			}
		}()
	}

	<-this.end
}
