package runtimes

import (
	"strings"
)

var defaultRt = map[string]*Runtime{
	"gnuc++11": Cpp11(),
	"gnuc++14": Cpp14(),
	"gnuc++17": Cpp17(),
	"gnuc++20": Cpp20(),
	"go":       Go(),
	"python3":  Py3(),
}

var Runtimes = buildRt()

type Runtime struct {
	getVersion  func() (string, error)
	Program     string `json:"program"`
	Arguments   string `json:"arguments"`
	ExecCommand string `json:"execCommand,omitempty"`
	Version     string `json:"version"`
}

func buildRt() (rt map[string]*Runtime) {
	rt = make(map[string]*Runtime)
	for key, r := range defaultRt {
		v, e := r.getVersion()
		if e != nil {
			continue
		}
		nr := r
		nr.Version = v
		rt[key] = nr
	}
	return
}

func (rt *Runtime) BuildCompileCommand(inp, output string) (string, []string) {
	r := strings.NewReplacer("{{input}}", inp, "{{output}}", output)
	return rt.Program, strings.Split(r.Replace(rt.Arguments), " ")
}

func (rt *Runtime) BuildExecCommand(prog string) (string, []string) {
	if rt.ExecCommand == "" {
		return prog, []string{}
	}
	c := strings.Split(strings.ReplaceAll(rt.ExecCommand, "{{program}}", prog), " ")
	return c[0], c[1:]
}
