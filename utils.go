package rest

import (
	"log"
	"os"
)

func IsVerbose() bool {
	if _, ok := os.LookupEnv("VERBOSE"); ok {
		return true
	} else {
		return false
	}
}

func LogVerbose(format string, data ...interface{}) {
	if IsVerbose() {
		log.Printf(format, data...)
	}
}
