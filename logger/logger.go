package logger

import (
	"github.com/rs/zerolog"
	"os"
	"runtime"
)

var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

func Panic(e error, msg string, args ...interface{}) {
	if e == nil {
		return
	}
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	l := Logger.Fatal().Err(e)
	if ok && details != nil {
		l = l.Str("from", details.Name())
	}
	l.Msgf(msg, args...)
}
