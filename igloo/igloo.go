package igloo

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.arsenm.dev/drpc/muxserver"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"igloo/igloo/judge/compiler"
	"igloo/igloo/pb"
	"igloo/igloo/sys"
	"log"
	"net"
	"os"
	"sort"
	"storj.io/drpc/drpcmux"
	"time"
)

var BootTimestamp = time.Now()

type Server struct {
	Specs pb.InstanceSpecification
}

var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})

func (c *Server) Specification(context.Context, *emptypb.Empty) (*pb.InstanceSpecification, error) {
	return &c.Specs, nil
}

func (c *Server) Alive(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return wrapperspb.Bool(true), nil
}

func (c *Server) Judge(file *pb.File, stream pb.DRPCIgloo_JudgeStream) error {
	file.Constraints = &pb.Constraints{
		Mem:      128,
		Cpu:      1,
		Duration: 1,
		Language: "cpp11",
	}
	if comp, ok := compiler.Compilers[file.Constraints.Language]; ok {
		comp.CompileAndRun(file)
	}
	return nil
}

func GetSpecs() (specs pb.InstanceSpecification) {
	specs = pb.InstanceSpecification{
		Os:            sys.GetOs(),
		Version:       "0.0.1-prealpha",
		BootTimestamp: timestamppb.New(BootTimestamp),
	}
	var runtimes []*pb.Runtime
	for key, comp := range compiler.Compilers {
		runtimes = append(runtimes, &pb.Runtime{
			Name:      comp.Name,
			Command:   comp.Command,
			Arguments: comp.Arguments,
			Key:       key,
			Version:   comp.Version,
		})
	}
	sort.Slice(runtimes, func(i, j int) bool {
		return runtimes[i].Key < runtimes[j].Key
	})
	specs.Runtimes = runtimes
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
	e := pb.DRPCRegisterIgloo(mux, &Server{
		Specs: GetSpecs(),
	})
	if e != nil {
		panic(e)
	}
	lis, err := net.Listen("tcp", ":1727")
	if err != nil {
		panic(err)
	}
	igloo := muxserver.New(mux)
	Logger.Info().Msgf("Listening on %s", ":1727")
	log.Fatalln(igloo.Serve(context.Background(), lis))
}
