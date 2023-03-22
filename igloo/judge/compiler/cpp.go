package compiler

import (
	"fmt"
	"igloo/igloo/utils"
	"strings"
)

var gccVersion = getGccVersion()

func Cpp20() *Compiler {
	return cpp("C++ 20", "c++20")
}

func Cpp17() *Compiler {
	return cpp("C++ 17", "c++17")
}

func Cpp14() *Compiler {
	return cpp("C++ 14", "c++14")
}

func Cpp11() *Compiler {
	return cpp("C++ 11", "c++11")
}

func getGccVersion() string {
	output, e := utils.Invoke("g++", "-dumpversion")
	if e != nil {
		return "unknown"
	}
	return strings.TrimSpace(output)
}

func cpp(name, std string) *Compiler {
	return &Compiler{
		Name:                name,
		Command:             "g++",
		TimeLimitMultiplier: 1,
		Arguments:           fmt.Sprintf("-std=%s -o {{output}} -Wall -DONLINE_JUDGE -O2 -lm -fmax-errors=5 -march=native -s {{input}}", std),
		Extensions:          []string{"cpp"},
		Version:             gccVersion,
	}
}
