package crawler

import (
	"github.com/nladuo/go-webcrawler/model/response"
	"github.com/nladuo/go-webcrawler/scheduler"
)

type Parser struct {
	Identifier string
	Parse      func(res *response.HttpResponse, processor *scheduler.Processor)
}
