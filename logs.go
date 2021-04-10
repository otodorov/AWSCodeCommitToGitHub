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
		// 0 = This function
		// 1 = Function that called this function
		_, fn, line, _ := runtime.Caller(1)
		fmt.Printf("DEBUG   | %s:%d | %v\n", fn, line, e)
	}
}
