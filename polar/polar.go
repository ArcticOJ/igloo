package polar

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.arsenm.dev/drpc/muxserver"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"log"
	"net"
	"polar/polar/pb"
	"polar/polar/sys"
	"storj.io/drpc/drpcmux"
	"time"
)

type Server struct {
	BootTimestamp time.Time
	Specs         pb.Specifications
}

func (c *Server) Health(context.Context, *emptypb.Empty) (*pb.PolarHealth, error) {
	return &pb.PolarHealth{
		Version:       "0.0.1-prealpha",
		BootTimestamp: timestamppb.New(c.BootTimestamp),
		Specs:         &c.Specs,
	}, nil
}

func (c *Server) IsAlive(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return wrapperspb.Bool(true), nil
}

func (c *Server) Judge(file *pb.File, stream pb.DRPCPolar_JudgeStream) error {
	return nil
}

func GetSpecs() (specs pb.Specifications) {
	specs = pb.Specifications{
		Os: sys.GetOs(),
	}
	if memInfo, e := mem.VirtualMemory(); e != nil {
		specs.Mem = -1
	} else {
		specs.Mem = float64(memInfo.Total) / (1 << 30)
	}
	if cpuInfo, e := cpu.Info(); e != nil || len(cpuInfo) == 0 {
		specs.Cpu = "Unknown"
	} else {
		specs.Cpu = cpuInfo[0].ModelName
	}
	return
}

func Start() {
	mux := drpcmux.New()
	e := pb.DRPCRegisterPolar(mux, &Server{
		BootTimestamp: time.Now(),
		Specs:         GetSpecs(),
	})
	if e != nil {
		panic(e)
	}
	lis, err := net.Listen("tcp", ":172")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Listening on tcp:%d", 172)
	polar := muxserver.New(mux)
	log.Fatalln(polar.Serve(context.Background(), lis))
}
