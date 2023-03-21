package compiler

import (
	"igloo/igloo/utils"
	"regexp"
)

var goVersion = getGoVersion()

var goVerPattern = regexp.MustCompile(`go version go(?P<Version>([0-9].[0-9]+(.[0-9]+)?))`)

func getGoVersion() string {
	output, e := utils.Invoke("go", "version")
	if e != nil {
		return "unknown"
	}
	version := goVerPattern.FindStringSubmatch(output)[goVerPattern.SubexpIndex("Version")]
	return version
}

func Go() *Compiler {
	return &Compiler{
		Name:       "Go",
		Command:    "go",
		Arguments:  "build -o %s %s",
		Extensions: []string{"go"},
		Version:    goVersion,
	}
}
