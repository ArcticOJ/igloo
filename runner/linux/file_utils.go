package runner

import (
	"os"
)

func prepareFiles(inp string, out string) ([]*os.File, error) {
	var err error
	files := make([]*os.File, 2)
	if inp != "" {
		files[0], err = os.OpenFile(inp, os.O_RDONLY, 0644)
		if err != nil {
			goto openerr
		}
	}
	if out != "" {
		//files[1] = os.Stdout
		files[1], err = os.OpenFile(out, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
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
