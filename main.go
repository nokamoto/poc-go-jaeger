package main

import (
	"flag"
	"fmt"
	pb "github.com/nokamoto/poc-go-jaeger/service"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

var port = flag.Int("port", 9090, "grpc server port")
var addr = flag.String("addr", "localhost:9090", "grpc client dial addr")
var collector = flag.String("collector", "http://jaeger:14268/api/traces", "jaeger exporter collector endpoint")

func serve() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(fmt.Sprintf("listen tcp port (%d) - %v", *port, err))
	}

	fmt.Printf("listen tcp port (%d)\n", *port)

	opts := []grpc.ServerOption{}
	opts = append(opts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	server := grpc.NewServer(opts...)

	a, err := newServiceA()
	if err != nil {
		panic(err)
	}

	pb.RegisterServiceAServer(server, a)
	pb.RegisterServiceBServer(server, &serviceB{})
	reflection.Register(server)

	fmt.Println("ready to serve")
	err = server.Serve(lis)
	if err != nil {
		panic(fmt.Sprintf("serve %v - %v", lis, err))
	}
}

func export() {
	je, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: *collector,
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
}

func call() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithStatsHandler(new(ocgrpc.ClientHandler)))

	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		panic(fmt.Sprintf("err: %s %v", *addr, err))
	}
	defer conn.Close()

	client := pb.NewServiceAClient(conn)

	for {
		fmt.Println("call")
		ctx, span := trace.StartSpan(context.Background(), "ClientCall")
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Duration(3000)*time.Millisecond))

		res, err := client.Send(ctx, &pb.Request{})
		if err != nil {
			span.SetStatus(trace.Status{Code: int32(grpc.Code(err)), Message: grpc.ErrorDesc(err)})
			fmt.Printf("err: %v\n", err)

			if grpc.Code(err) == codes.DeadlineExceeded {
				cancel()
			}
		} else {
			fmt.Printf("rec: %v\n", res)
		}
		span.End()

		time.Sleep(time.Duration(5000) * time.Millisecond)
	}
}

func main() {
	flag.Parse()

	export()

	go call()

	serve()
}
