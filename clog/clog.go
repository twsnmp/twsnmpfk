package clog

import (
	"log"

	"github.com/fatih/color"
)

type ColorLogger struct{}

func New() *ColorLogger {
	return &ColorLogger{}
}

func (l *ColorLogger) Print(message string) {
	log.Println(message)
}

func (l *ColorLogger) Trace(message string) {
	log.Printf("%s %s", color.CyanString("[TRACE]"), message)
}

func (l *ColorLogger) Debug(message string) {
	log.Printf("%s %s", color.HiGreenString("[DEBUG]"), message)
}

func (l *ColorLogger) Info(message string) {
	log.Printf("%s %s", color.HiBlueString("[INFO]"), message)
}

func (l *ColorLogger) Warning(message string) {
	log.Printf("%s %s", color.HiYellowString("[WARN]"), message)
}

func (l *ColorLogger) Error(message string) {
	log.Printf("%s %s", color.HiMagentaString("[ERR]"), message)
}

func (l *ColorLogger) Fatal(message string) {
	log.Printf("%s %s", color.HiRedString("[FATAL]"), message)
}
