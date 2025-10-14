package app

import (
	"log"
	"os"
)

var (
    InfoLog  = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
    ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime)
)