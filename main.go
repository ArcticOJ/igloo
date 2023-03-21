package main

import (
	"igloo/igloo"
	"log"
	"runtime"
)

func init() {
	if runtime.GOOS != "linux" {
		log.Fatalln("Unsupported platform. Try using Docker instead.")
	}
}

func main() {
	igloo.Start()
}
