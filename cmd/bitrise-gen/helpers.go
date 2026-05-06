package main

import (
	"fmt"
	"os"
)

// fatalf prints a formatted error message to stderr and exits with code 1.
func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
