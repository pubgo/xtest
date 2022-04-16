package model

type Result struct {
	Service  string
	Name     string
	Request  interface{}
	Response interface{}
	Error    string
	Status   bool
}
