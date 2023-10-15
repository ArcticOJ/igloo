package main

import (
	"fmt"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
)

func main() {
	rt := make(map[string]runtimes.Runtime)
	for name, r := range runtimes.DefaultSupportedRuntimes {
		_r := r
		if ver, e := _r.GetVersion(); e == nil {
			p, _e := exec.LookPath(r.Program)
			if _e != nil {
				fmt.Printf("'%s' does not exist in PATHS, skipping\n", r.Program)
			}
			rt[name] = runtimes.Runtime{
				Program:     p,
				Arguments:   _r.Arguments,
				ExecCommand: _r.ExecCommand,
				Version:     ver,
			}
			continue
		}
		fmt.Printf("'%s' is not available, skipping\n", name)
	}
	buf, e := yaml.Marshal(rt)
	if e != nil {
		panic(e)
	}
	if e := os.WriteFile("runtimes.yml", buf, 0755); e != nil {
		panic(e)
	}
}
