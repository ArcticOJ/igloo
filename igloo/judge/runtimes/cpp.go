package definitions

import (
	"fmt"
	"igloo/igloo/judge/runtimes"
	"igloo/igloo/utils"
	"strings"
)

var gccVersion = getGccVersion()

func Cpp20() *runtimes.Runtime {
	return cpp("C++ 20", "c++20")
}

func Cpp17() *runtimes.Runtime {
	return cpp("C++ 17", "c++17")
}

func Cpp14() *runtimes.Runtime {
	return cpp("C++ 14", "c++14")
}

func Cpp11() *runtimes.Runtime {
	return cpp("C++ 11", "c++11")
}

func getGccVersion() string {
	output, e := utils.Invoke("g++", "-dumpversion")
	if e != nil {
		return "unknown"
	}
	return strings.TrimSpace(output)
}

func cpp(name, std string) *runtimes.Runtime {
	return &runtimes.Runtime{
		Name:                name,
		Command:             "g++",
		TimeLimitMultiplier: 1,
		Arguments:           fmt.Sprintf("-std=%s -o {{output}} -Wall -DONLINE_JUDGE -O2 -lm -fmax-errors=5 -march=native -s {{input}}", std),
		Extension:           "cpp",
		Version:             gccVersion,
	}
}
