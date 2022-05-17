package utils

import (
	"log"
	"time"
)

// ExecutionTime
func ExecutionTime(start time.Time, name string) {
	elapsed := time.Since(start)

	log.Printf("%s took %s", name, elapsed)
}
