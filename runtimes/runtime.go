package runtimes

import (
	"strings"
)

var (
	DefaultSupportedRuntimes = map[string]supportedRuntime{
		"gnuc++11": gnucpp11(),
		"gnuc++14": gnucpp14(),
		"gnuc++17": gnucpp17(),
		"gnuc++20": gnucpp20(),
		"golang":   golang(),
		"python3":  py3(),
	}
)

type (
	supportedRuntime struct {
		GetVersion  func() (string, error)
		Program     string
		Arguments   string
		ExecCommand string
	}
	Runtime struct {
		Program     string `yaml:"program"`
		Arguments   string `yaml:"arguments"`
		ExecCommand string `yaml:"execCommand,omitempty"`
		Version     string `yaml:"version"`
	}
)

func (rt Runtime) BuildCompileCommand(inp, output string) (string, []string) {
	r := strings.NewReplacer("{{input}}", inp, "{{output}}", output)
	return rt.Program, strings.Split(r.Replace(rt.Arguments), " ")
}

func (rt Runtime) BuildExecCommand(prog string) (string, []string) {
	if rt.ExecCommand == "" {
		return prog, []string{}
	}
	c := strings.Split(strings.ReplaceAll(rt.ExecCommand, "{{program}}", prog), " ")
	return c[0], c[1:]
}
