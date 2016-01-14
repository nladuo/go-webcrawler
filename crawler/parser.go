package crawler

import (
	"github.com/nladuo/go-webcrawler/model/result"
	"github.com/nladuo/go-webcrawler/scheduler"
)

type Parser struct {
	Identifier string
	Parse      func(res *result.Result, processor scheduler.Processor)
}
