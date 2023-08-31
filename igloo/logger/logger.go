package logger

import (
	"github.com/rs/zerolog"
	"os"
)

var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

func Panic(e error, msg string, args ...interface{}) {
	if e == nil {
		return
	}
	Logger.Fatal().Err(e).Msgf(msg, args...)
}
