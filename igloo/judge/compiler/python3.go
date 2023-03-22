package compiler

import (
	"igloo/igloo/utils"
	"regexp"
)

var py3Version = getPy3Version()

var py3VerPattern = regexp.MustCompile(`Python (?P<Version>([0-9].[0-9]+(.[0-9]+)?))`)

func getPy3Version() string {
	output, e := utils.Invoke("python3", "--version")
	if e != nil {
		return "unknown"
	}
	version := py3VerPattern.FindStringSubmatch(output)[goVerPattern.SubexpIndex("Version")]
	return version
}

func Py3() *Compiler {
	return &Compiler{
		Name:       "Python 3",
		Command:    "python3",
		Arguments:  "-I -B {{input}}",
		Extensions: []string{"py"},
		Version:    py3Version,
	}
}
