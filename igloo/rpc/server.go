package server

import (
	"context"
	"go.arsenm.dev/drpc/muxserver"
	"igloo/igloo/global"
	igloopb "igloo/igloo/pb"
	"log"
	"net"
	"storj.io/drpc/drpcmux"
)

func Start() {
	mux := drpcmux.New()
	e := igloopb.DRPCRegisterIgloo(mux, new(Server))
	if e != nil {
		panic(e)
	}
	lis, err := net.Listen("tcp", ":1727")
	if err != nil {
		panic(err)
	}
	server := muxserver.New(createMiddleware("test")(mux))
	global.Logger.Info().Msgf("Listening on %s", ":1727")
	log.Fatalln(server.Serve(context.Background(), lis))
}
