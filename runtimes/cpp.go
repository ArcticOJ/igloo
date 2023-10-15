package runtimes

import (
	"fmt"
	"github.com/ArcticOJ/igloo/v0/utils"
	"strings"
)

func gnucpp20() supportedRuntime {
	return cpp("c++20")
}

func gnucpp17() supportedRuntime {
	return cpp("c++17")
}

func gnucpp14() supportedRuntime {
	return cpp("c++14")
}

func gnucpp11() supportedRuntime {
	return cpp("c++11")
}

func getGccVersion() (string, error) {
	output, e := utils.InvokeStdout("g++", "-dumpversion")
	if e != nil {
		return "", e
	}
	return strings.TrimSpace(output), nil
}

func cpp(std string) supportedRuntime {
	return supportedRuntime{
		Program:    "g++",
		Arguments:  fmt.Sprintf("-std=%s -o {{output}} -Wall -DONLINE_JUDGE -O2 -lm -march=native -s {{input}}", std),
		GetVersion: getGccVersion,
	}
}
