package crawler

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nladuo/DLocker"
	"github.com/nladuo/go-webcrawler/downloader"
	"github.com/nladuo/go-webcrawler/model"
	"github.com/nladuo/go-webcrawler/scheduler"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

const (
	ErrShutDownCrawler string = "Cannot ShutDown the crwaler when "
)

// arguments for zookeeper
const (
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
	threadNum     int
	parsers       []*model.Parser
	scheduler     scheduler.Scheduler
	processor     model.Processor
	downloader    downloader.Downloader
	isMaster      bool      //only the master crawler excute the Crawler.AddBaseTask
	end           chan byte //unbufferred channel
	threadManager *model.ThreadManager
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
	sqlScheduler := scheduler.NewDistributedSqlScheduler(db, lockersPath, lockerTimeout)
	crawler.scheduler = sqlScheduler
	crawler.processor = sqlScheduler

	//make unbufferred channel
	crawler.end = make(chan byte, 0)

	//set the downloader to nil
	crawler.downloader = nil
	//set the thread_manager to nil
	crawler.threadManager = nil
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

	//make unbufferred channel
	crawler.end = make(chan byte, 0)

	//set the downloader to nil
	crawler.downloader = nil
	//set the thread_manager to nil
	crawler.threadManager = nil
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

	//make unbufferred channel
	crawler.end = make(chan byte, 0)

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

func (this *Crawler) SetProxyTimeOut(timeout time.Duration) {
	downloader.SetProxyTimeOut(timeout)
}

func (this *Crawler) Run() {
	if this.downloader == nil {
		this.downloader = downloader.NewDefaultDownloader()
	}

	this.threadManager = model.NewThreadManager(this.threadNum)

	log.Println("Crawler start running....")

	for {
		task := this.scheduler.GetTask()
		tag_int := this.threadManager.GetOccupation()
		tag := fmt.Sprintf("[goroutine %d]", tag_int)
		go func(tag string, task model.Task) { //async download task
			result := this.downloader.Download(tag, task)
			if result.Err != nil {
				this.threadManager.FreeOccupation()
				return
			}
			log.Println("Get Result, Identifier: " + result.Identifier)
			for i := 0; i < len(this.parsers); i++ {
				if this.parsers[i].Identifier == result.Identifier {
					this.parsers[i].Parse(result, this.processor)
					break
				}
			}
			this.threadManager.FreeOccupation()
		}(tag, task)

	}
}

func (this *Crawler) ShutDown() {
	if this.threadManager == nil {
		panic(errors.New(ErrShutDownCrawler))
	}
	this.threadManager.GetOccupation()
	time.Sleep(100 * time.Millisecond)
	if this.scheduler.GetTaskSize() == 0 {
		// if there is no task in taskchan, shutdown the crawler
		log.Println("Crawler has finished....")
		os.Exit(0)
	}
	this.threadManager.FreeOccupation()

}

//for debug
func (this *Crawler) SetPProfPort(port string) {
	go func() {
		log.Println(http.ListenAndServe("localhost:"+port, nil))
	}()
}
