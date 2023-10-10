package runtimes

import (
	"github.com/ArcticOJ/igloo/v0/utils"
)

var py3VerRegex = utils.NewRegex(`Python (?P<Version>([0-9].[0-9]+(.[0-9]+)?))`)

func getPy3Version() (string, error) {
	output, e := utils.InvokeStdout("python3", "--version")
	if e != nil {
		return "", e
	}
	return py3VerRegex.Submatch(output).Find("Version"), nil
}

func Py3() *Runtime {
	return &Runtime{
		Program:     "python3",
		Arguments:   "-m libarctic.compiler -q -o {{output}}.pyc {{input}}",
		ExecCommand: "python3 {{program}}.pyc",
		getVersion:  getPy3Version,
	}
}
