package main

import (
	"context"
	"fmt"
	"hello_server/pb"
	"net"

	"google.golang.org/grpc"
)

// grpc server
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello 是我们需要实现的方法
// 这个方法是我们对外提供的服务
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	reply := "hello " + in.GetName()
	return &pb.HelloResponse{
		Reply: reply,
	}, nil
}

func (s *server) Add(ctx context.Context, param *pb.AddParam) (*pb.Result, error) {
	result := int64(param.X) + int64(param.Y)
	return &pb.Result{
		Z: result,
	}, nil
}

func main() {
	// 启动服务
	l, err := net.Listen("tcp", ":8972")
	if err != nil {
		fmt.Printf("fail to listen, err: %v\n", err)
	}
	s := grpc.NewServer() // 创建grpc服务
	// 注册服务
	pb.RegisterGreeterServer(s, &server{})
	// 启动服务
	if err = s.Serve(l); err != nil {
		fmt.Printf("failed to serve , err: %v\n", err)
		return
	}
}
