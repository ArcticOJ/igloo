package runtimes

import (
	"fmt"
	"igloo/igloo/utils"
	"strings"
)

func Cpp20() *Runtime {
	return cpp("c++20")
}

func Cpp17() *Runtime {
	return cpp("c++17")
}

func Cpp14() *Runtime {
	return cpp("c++14")
}

func Cpp11() *Runtime {
	return cpp("c++11")
}

func getGccVersion() (string, error) {
	output, e := utils.InvokeStdout("g++", "-dumpversion")
	if e != nil {
		return "", e
	}
	return strings.TrimSpace(output), nil
}

func cpp(std string) *Runtime {
	return &Runtime{
		Program:    "g++",
		Arguments:  fmt.Sprintf("-std=%s -o {{output}} -Wall -DONLINE_JUDGE -O2 -lm -march=native -s {{input}}", std),
		getVersion: getGccVersion,
	}
}
