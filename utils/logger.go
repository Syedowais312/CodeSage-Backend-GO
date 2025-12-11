package utils

import (
    "log"
)

// Infof logs informational messages with a consistent prefix.
func Infof(format string, v ...any) {
    log.Printf("[INFO] "+format, v...)
}

// Errorf logs error messages with a consistent prefix.
func Errorf(format string, v ...any) {
    log.Printf("[ERROR] "+format, v...)
}