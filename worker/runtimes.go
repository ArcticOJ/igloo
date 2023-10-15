package worker

import (
	"github.com/ArcticOJ/igloo/v0/logger"
	"github.com/ArcticOJ/igloo/v0/runtimes"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
)

var Runtimes = getRuntimes()

func getRuntimes() (rt map[string]runtimes.Runtime) {
	rt = make(map[string]runtimes.Runtime)
	buf, e := os.ReadFile("runtimes.yml")
	logger.Panic(e, "could not read 'runtimes.yml'")
	logger.Panic(yaml.Unmarshal(buf, &rt), "could not parse runtimes list")
	for name, r := range rt {
		if _, ok := runtimes.DefaultSupportedRuntimes[name]; !ok {
			delete(rt, name)
			logger.Logger.Warn().Msgf("'%s' is not supported, skipping", name)
		}
		if _, e := exec.LookPath(r.Program); e != nil {
			delete(rt, name)
			logger.Logger.Warn().Msgf("'%s' is not available, skipping", r.Program)
		}
	}
	if len(rt) == 0 {
		logger.Logger.Fatal().Msg("no available runtime, exiting")
	}
	return
}
