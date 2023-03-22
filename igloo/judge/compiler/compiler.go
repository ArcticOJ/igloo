package compiler

import (
	_ "embed"
	"fmt"
	"igloo/igloo/judge"
	"igloo/igloo/pb"
	"igloo/igloo/utils"
	"os"
	"strings"
)

var Compilers = map[string]*Compiler{
	"cpp11":   Cpp11(),
	"cpp14":   Cpp14(),
	"cpp17":   Cpp17(),
	"cpp20":   Cpp20(),
	"go":      Go(),
	"python3": Py3(),
}

type Compiler struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Extensions []string `json:"extensions"`
	Command    string   `json:"program"`
	Arguments  string   `json:"arguments"`
}

func (compiler *Compiler) BuildCommand(inp, output string) (string, []string) {
	r := strings.NewReplacer("{{input}}", inp, "{{output}}", output)
	return compiler.Command, strings.Split(r.Replace(compiler.Arguments), " ")
}

func (compiler *Compiler) CompileAndRun(file *pb.File) {
	targetOut := fmt.Sprintf("/tmp/%s", file.Id)
	fmt.Println(os.WriteFile(targetOut+".cpp", file.Buffer, 0755))
	cmd, args := compiler.BuildCommand(targetOut+".cpp", targetOut)
	fmt.Println(utils.Invoke(cmd, args...))
	fmt.Println(judge.Start(&judge.Config{
		IOFileName:    "/tmp/HELLOWORLD",
		MemoryLimit:   file.Constraints.Mem << 20,
		TimeLimit:     file.Constraints.Duration,
		TimeLimitHard: file.Constraints.Duration + 2,
		StackLimit:    1024 << 20,
		OutputLimit:   64 << 20,
		Verbose:       true,
		Type:          judge.Cpp,
	}, []string{targetOut}))
}
