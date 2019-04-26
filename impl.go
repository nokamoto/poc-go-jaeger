package main

import (
	"fmt"
	pb "github.com/nokamoto/poc-go-jaeger/service"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"time"
)

type serviceA struct {
	cli pb.ServiceBClient
}

func newServiceA() (*serviceA, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithStatsHandler(new(ocgrpc.ClientHandler)))

	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		return nil, err
	}

	return &serviceA{cli: pb.NewServiceBClient(conn)}, nil
}

func (a *serviceA) Send(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	ctx, span := trace.StartSpan(ctx, "ServiceA.Send")
	defer span.End()

	res, err := a.cli.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	milli := rand.Int63n(3000)
	fmt.Printf("start: heavy workload A %d millis\n", milli)
	time.Sleep(time.Duration(milli) * time.Millisecond)

	if ctx.Err() == context.Canceled {
		fmt.Println("A: Client cancelled, abandoning.")
		return nil, status.Error(codes.Canceled, "A: Client cancelled, abandoning.")
	}

	fmt.Printf("done: heavy workload A %d millis\n", milli)

	return res, nil
}

type serviceB struct{}

func (*serviceB) Send(ctx context.Context, _ *pb.Request) (*pb.Response, error) {
	ctx, span := trace.StartSpan(ctx, "ServiceB.Send")
	defer span.End()

	milli := rand.Int63n(6000)
	fmt.Printf("start: heavy workload B %d millis\n", milli)
	time.Sleep(time.Duration(milli) * time.Millisecond)

	if ctx.Err() == context.Canceled {
		fmt.Println("B: Client cancelled, abandoning.")
		return nil, status.Error(codes.Canceled, "B: Client cancelled, abandoning.")
	}

	fmt.Printf("done: heavy workload B %d millis\n", milli)

	return &pb.Response{}, nil
}
