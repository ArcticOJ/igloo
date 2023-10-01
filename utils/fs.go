package utils

import (
	"errors"
	"os"
	"path"
)

func Clean(files ...string) {
	for _, file := range files {
		os.Remove(file)
	}
}

func CreateRandomFile(prefix string) (string, error) {
	tmp := os.TempDir()
	for try := 0; try < 10; try++ {
		f := path.Join(tmp, prefix+NextRand())
		_, e := os.Stat(f)
		if e != nil && os.IsNotExist(e) {
			return f, nil
		}
	}
	return "", errors.New("failed to create a random file")
}
