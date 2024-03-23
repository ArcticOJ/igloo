package build

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

var (
	Version       = "n/a"
	Hash          = "n/a"
	_date         = "0"
	Date    int64 = 0
	Tag           = "dev"
)

func init() {
	d, e := strconv.Atoi(_date)
	if e != nil {
		return
	}
	Date = int64(d)
	Version = Tag
	if Tag == "dev" {
		Version = fmt.Sprintf("%s#%s@%s", Tag, Hash, time.Unix(Date, 0).Format(time.RFC3339))
	}
	Version = fmt.Sprintf("%s with %s", Version, runtime.Version())
}
