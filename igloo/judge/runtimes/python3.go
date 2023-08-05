package definitions

import (
	"igloo/igloo/judge/runtimes"
	"igloo/igloo/utils"
)

var py3Version = getPy3Version()

var py3VerRegex = utils.NewRegex(`Python (?P<Version>([0-9].[0-9]+(.[0-9]+)?))`)

func getPy3Version() string {
	output, e := utils.Invoke("python3", "--version")
	if e != nil {
		return "unknown"
	}
	return py3VerRegex.Submatch(output).Find("Version")
}

func Py3() *runtimes.Runtime {
	return &runtimes.Runtime{
		Name:                "Python 3",
		Command:             "python3",
		TimeLimitMultiplier: 1,
		Arguments:           "-m compileall -q -I {{input}}",
		Extension:           "py",
		Version:             py3Version,
	}
}
