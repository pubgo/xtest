package xtest

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[xtest] ", log.LstdFlags|log.Lshortfile)

// Log ...
func Log() *log.Logger {
	return logger
}
