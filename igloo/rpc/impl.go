package server

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"igloo/igloo/global"
	igloopb "igloo/igloo/pb"
)

type Server struct{}

func (*Server) Specification(context.Context, *emptypb.Empty) (*igloopb.InstanceSpecification, error) {
	return &global.Specs, nil
}

func (*Server) Alive(context.Context, *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return wrapperspb.Bool(true), nil
}

func (*Server) Judge(sub *igloopb.Submission, stream igloopb.DRPCIgloo_JudgeStream) error {
	defer global.Worker.Unsubscribe(sub.Id)
	res := make(chan *igloopb.JudgeResult)
	global.Worker.Subscribe(sub.Id, res)
	global.Queue.Enqueue(sub)
	for r := range res {
		stream.Send(r)
		if _, ok := r.Result.(*igloopb.JudgeResult_Final); ok {
			break
		}
	}
	return stream.Close()
}
