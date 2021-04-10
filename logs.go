package main

import (
	"fmt"
	"runtime"
)

// Handle errors
func logHandler(t, e string) {
	switch t {
	case "warning":
		fmt.Println("WARNING |", e)
	case "error":
		fmt.Println("ERROR   |", e)
	case "debug":
		_, fn, line, _ := runtime.Caller(1)
		fmt.Printf("DEBUG   | %s:%d | %v\n", fn, line, e)
	}
}
