package definitions

import (
	"igloo/igloo/judge/runtimes"
	"igloo/igloo/utils"
)

var goVersion = getGoVersion()

var goVerRegex = utils.NewRegex(`go version go(?P<Version>([0-9].[0-9]+(.[0-9]+)?))`)

func getGoVersion() string {
	output, e := utils.Invoke("go", "version")
	if e != nil {
		return "unknown"
	}
	return goVerRegex.Submatch(output).Find("Version")
}

func Go() *runtimes.Runtime {
	return &runtimes.Runtime{
		Name:                "Go",
		Command:             "go",
		TimeLimitMultiplier: 1,
		Arguments:           "build -o {{input}} {{output}}",
		Extension:           "go",
		Version:             goVersion,
	}
}
