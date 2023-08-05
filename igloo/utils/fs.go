package utils

import "os"

func Clean(files ...string) {
	for _, file := range files {
		os.Remove(file)
	}
}
