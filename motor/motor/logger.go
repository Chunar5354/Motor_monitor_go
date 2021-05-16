package motor

import (
	"log"
	"io"
	"os"
)

var (
	Info *log.Logger
	Error *log.Logger
)

func LogInit() {
	file, err := os.OpenFile("/home/Chunar/codes/motor/motor.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	Info = log.New(io.MultiWriter(file, os.Stderr), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(file, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
