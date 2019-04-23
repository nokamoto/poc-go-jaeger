package main

import (
	"flag"
	"fmt"
	pb "github.com/nokamoto/poc-go-jaeger/service"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

type serviceA struct{}

func (*serviceA) Send(_ context.Context, _ *pb.Request) (*pb.Response, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "not implemented yet")
}

type serviceB struct{}

func (*serviceB) Send(_ context.Context, _ *pb.Request) (*pb.Response, error) {
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
			ctx := context.Background()
			res, err := client.Send(ctx, &pb.Request{})
			if err != nil {
				fmt.Printf("err: %v\n", err)
			} else {
				fmt.Printf("rec: %v\n", res)
			}

			time.Sleep(3 * time.Second)
		}
	}()

	fmt.Println("ready to serve")
	err = server.Serve(lis)
	if err != nil {
		panic(fmt.Sprintf("serve %v - %v", lis, err))
	}
}
