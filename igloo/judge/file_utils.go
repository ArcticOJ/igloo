package judge

import (
	"github.com/criyle/go-sandbox/pkg/memfd"
	"io"
	"os"
)

func prepareFiles(input io.Reader, outputFile, errorFile string) ([]*os.File, error) {
	var err error
	files := make([]*os.File, 3)
	if input != nil {
		files[0], err = memfd.DupToMemfd("input", input)
		if err != nil {
			goto openerr
		}
	}
	if outputFile != "" {
		files[1], err = os.OpenFile(outputFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			goto openerr
		}
	}
	if errorFile != "" {
		files[2], err = os.OpenFile(errorFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			goto openerr
		}
	}
	return files, nil
openerr:
	closeFiles(files)
	return nil, err
}

func closeFiles(files []*os.File) {
	for _, f := range files {
		if f != nil {
			f.Close()
		}
	}
}
