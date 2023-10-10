package logger

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"os"
	"runtime"
	"time"
)

var Logger zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, FormatTimestamp: func(i interface{}) string {
		t := i.(json.Number)
		if _t, e := t.Int64(); e != nil {
			return ""
		} else {
			return "\033[0;36m " + time.UnixMilli(_t).Format("02/01/2006 15:04:05.000") + "\033[0m"
		}
	}}).With().Timestamp().Logger()
}

func Panic(e error, msg string, args ...interface{}) {
	if e == nil {
		return
	}
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	l := Logger.Panic().Err(e)
	if ok && details != nil {
		l = l.Str("from", details.Name())
	}
	l.Msgf(msg, args...)
}
