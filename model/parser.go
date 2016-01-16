package model

type Parser struct {
	Identifier string
	Parse      func(res *Result, processor Processor)
}
