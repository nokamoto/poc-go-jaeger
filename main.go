package main

import (
	"flag"
	"fmt"
	pb "github.com/nokamoto/poc-go-jaeger/service"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type serviceA struct{}

func (*serviceA) Send(ctx context.Context, _ *pb.Request) (*pb.Response, error) {
	ctx, span := trace.StartSpan(ctx, "ServiceA.Send")
	defer span.End()

	return nil, grpc.Errorf(codes.Unimplemented, "not implemented yet")
}

type serviceB struct{}

func (*serviceB) Send(ctx context.Context, _ *pb.Request) (*pb.Response, error) {
	ctx, span := trace.StartSpan(ctx, "ServiceB.Send")
	defer span.End()

	return nil, grpc.Errorf(codes.Unimplemented, "not implemented yet")
}

func main() {
	port := flag.Int("port", 9090, "grpc server port")
	addr := flag.String("addr", "localhost:9090", "grpc client dial addr")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(fmt.Sprintf("listen tcp port (%d) - %v", *port, err))
	}

	fmt.Printf("listen tcp port (%d)\n", *port)

	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)

	pb.RegisterServiceAServer(server, &serviceA{})
	pb.RegisterServiceBServer(server, &serviceB{})
	reflection.Register(server)

	go func() {
		time.Sleep(3 * time.Second)

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())

		conn, err := grpc.Dial(*addr, opts...)
		if err != nil {
			panic(fmt.Sprintf("err: %s %v", *addr, err))
		}
		defer conn.Close()

		client := pb.NewServiceAClient(conn)

		for {
			ctx, span := trace.StartSpan(context.Background(), "ClientCall")
			res, err := client.Send(ctx, &pb.Request{})
			if err != nil {
				fmt.Printf("err: %v\n", err)
			} else {
				fmt.Printf("rec: %v\n", res)
			}
			span.End()

			time.Sleep(3 * time.Second)
		}
	}()

	agentEndpointURI := "jaeger:6831"
	collectorEndpointURI := "http://jaeger:14268/api/traces"

	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     agentEndpointURI,
		CollectorEndpoint: collectorEndpointURI,
		Process: jaeger.Process{
			ServiceName: "poc-go-jaeger",
		},
		OnError: func(err error) {
			fmt.Printf("err: %v\n", err)
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create the Jaeger exporter: %v", err))
	}
	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	fmt.Println("ready to serve")
	err = server.Serve(lis)
	if err != nil {
		panic(fmt.Sprintf("serve %v - %v", lis, err))
	}
}
