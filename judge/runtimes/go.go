package runtimes

import (
	"igloo/utils"
)

var goVerRegex = utils.NewRegex(`go version go(?P<Version>([0-9].[0-9]+(.[0-9]+)?))`)

func getGoVersion() (string, error) {
	output, e := utils.InvokeStdout("go", "version")
	if e != nil {
		return "", e
	}
	return goVerRegex.Submatch(output).Find("Version"), nil
}

func Go() *Runtime {
	return &Runtime{
		Program:    "go",
		Arguments:  "build -x -o {{output}} {{input}}",
		getVersion: getGoVersion,
	}
}
