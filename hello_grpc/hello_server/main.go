package main

import (
	"context"
	"fmt"
	"hello_server/pb"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// grpc server
type server struct {
	pb.UnimplementedGreeterServer
	mu    sync.Mutex     // count的并发锁
	count map[string]int // 存储每个name调用sayHello的次数
}

// SayHello 是我们需要实现的方法
// 这个方法是我们对外提供的服务
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	// 利用defer 在发送完响应数据后发送trailer
	defer func() {
		trailer := metadata.Pairs(
			"timestamp", strconv.Itoa(int(time.Now().Unix())),
		)
		grpc.SetTrailer(ctx, trailer)
	}()

	// 在执行业务逻辑之前要check metadata中是否包含token
	md, ok := metadata.FromIncomingContext(ctx)
	fmt.Printf("md:%#v ok:%#v\n", md, ok)

	if !ok {
		return nil, status.Error(codes.Unauthenticated, "无效请求!")
	}

	vl := md.Get("token")
	if len(vl) < 1 || vl[0] != "app-test-tang" {
		return nil, status.Error(codes.Unauthenticated, "无效请求!")
	}
	// if vl , ok := md["token"] ; ok {
	// 	if len(vl) > 0 && vl[0] == "app-test-tang" {
	// 		// 有效的请求

	// 	}
	// }

	s.mu.Lock()
	defer s.mu.Unlock()
	s.count[in.GetName()]++
	if s.count[in.GetName()] > 1 {
		// 返回请求次数限制的错误
		st := status.New(codes.ResourceExhausted, "请求次数限制")
		// 添加错误详情信息 需要接收返回的status
		ds, err := st.WithDetails(
			&errdetails.QuotaFailure{
				Violations: []*errdetails.QuotaFailure_Violation{
					{
						Subject:     fmt.Sprintf("name:%s", in.Name),
						Description: "每个name只能调用一次",
					},
				},
			},
		)
		if err != nil { // withDetails执行失败,返回原来的status.Err
			return nil, st.Err()
		}
		return nil, ds.Err()
	}

	reply := "hello " + in.GetName()

	// 发送数据前发送header
	header := metadata.New(map[string]string{
		"location": "Xian",
	})
	grpc.SendHeader(ctx, header)

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

// LotsOfReplies 返回使用多种语言打招呼
func (s *server) LotsOfReqlies(in *pb.HelloRequest, stream pb.Greeter_LotsOfReqliesServer) error {
	words := []string{
		"你好",
		"hello",
		"こんにちは",
		"안녕하세요",
	}

	for _, word := range words {
		data := &pb.HelloResponse{
			Reply: word + " " + in.GetName(),
		}
		// 使用Send方法返回多个数据
		if err := stream.Send(data); err != nil {
			return err
		}
	}
	return nil
}

// LotsOfGreetings 接收流式数据
func (s *server) LotsOfGreetings(stream pb.Greeter_LotsOfGreetingsServer) error {
	reply := "你好"
	for {
		// 接收客户端发来的流式数据
		res, err := stream.Recv()
		if err == io.EOF {
			// 最终统一回复
			return stream.SendAndClose(&pb.HelloResponse{
				Reply: reply,
			})
		}
		if err != nil {
			return err
		}
		reply += res.GetName()
	}
}

// BidiHello 双向流式打招呼
func (s *server) BidiHello(stream pb.Greeter_BidiHelloServer) error {
	for {
		// 接收流式请求
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		reply := magic(in.GetName()) // 对收到的数据做些处理

		// 返回流式响应
		if err := stream.Send(&pb.HelloResponse{Reply: reply}); err != nil {
			return err
		}
	}
}

func magic(s string) string {
	s = strings.ReplaceAll(s, "吗", "")
	s = strings.ReplaceAll(s, "吧", "")
	s = strings.ReplaceAll(s, "你", "我")
	s = strings.ReplaceAll(s, "？", "!")
	s = strings.ReplaceAll(s, "?", "!")
	return s
}

func main() {
	// 启动服务
	l, err := net.Listen("tcp", ":8972")
	if err != nil {
		fmt.Printf("fail to listen, err: %v\n", err)
	}

	// 加载证书信息
	creds , err := credentials.NewServerTLSFromFile("certs/server.crt", "certs/server.key")
	if err != nil {
		fmt.Printf("fail to load cert, err: %v\n", err)
		return
	}

	s := grpc.NewServer(grpc.Creds(creds)) // 创建grpc服务
	// 注册服务
	pb.RegisterGreeterServer(s, &server{count: make(map[string]int)})
	// 启动服务
	if err = s.Serve(l); err != nil {
		fmt.Printf("failed to serve , err: %v\n", err)
		return
	}
}
