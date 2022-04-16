package model

type Result struct {
	Id       int
	Service  string
	Name     string
	Request  interface{}
	Response interface{}
	Error    string
	Status   bool
}
