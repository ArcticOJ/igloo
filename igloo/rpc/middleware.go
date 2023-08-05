package server

import (
	"fmt"
	"igloo/igloo/global"
	"runtime/debug"
	"storj.io/drpc"
)

func createMiddleware(key string) func(handler drpc.Handler) drpc.Handler {
	return func(next drpc.Handler) drpc.Handler {
		return middleware{
			key:  key,
			next: next,
		}
	}
}

type middleware struct {
	next drpc.Handler
	key  string
}

func (mw middleware) HandleRPC(stream drpc.Stream, rpc string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%s", v)
			global.Logger.Printf("Recovered from panic: %v\n%s", v, debug.Stack())
		}
	}()
	/*data, ok := drpcmetadata.Get(stream.Context())
	if !ok {
		err = errors.New("missing auth key")
		fmt.Println("missing")
		return
	}
	if key, ok := data["key"]; !ok {
		err = errors.New("missing auth key")
		fmt.Println("missing")
		return
	} else if key != mw.key {
		err = errors.New("auth key mismatch")
		fmt.Println("mismatch")
		return
	}*/
	err = mw.next.HandleRPC(stream, rpc)
	return
}
