package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"hello_client/pb"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// grpc 客户端
// 调用server端的 SayHello 方法

var name = flag.String("name", "tang", "通过-name告诉server你是谁")

// var paramX = flag.Int("paramX" , 1, "通过-paramX输入paramX")
// var paramY = flag.Int("paramY" , 2, "通过-paramY输入paramY")

func main() {
	flag.Parse() // 解析命令行参数

	// 连接server
	conn, err := grpc.Dial("127.0.0.1:8972", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc.Dial failed,err:%v", err)
		return
	}
	defer conn.Close()
	// 创建客户端
	c := pb.NewGreeterClient(conn) // 使用生成的Go代码

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 调用RPC方法，普通rpc调用
	// 带元数据
	md := metadata.Pairs(
		"token", "app-test-tang",
	)
	ctx = metadata.NewOutgoingContext(ctx, md)
	var header, trailer metadata.MD
	respHello, err := c.SayHello(
		ctx,
		&pb.HelloRequest{Name: *name},
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)
	if err != nil {
		log.Printf("c.SayHellofailed, err:%v", err)
		return
	}

	// 拿到响应数据之前,获取header
	fmt.Printf("header:%v\n", header)

	// 拿到了RPC响应
	log.Printf("resp:%v", respHello.GetReply())
	// log.Printf("resp:%v", respAdd.GetZ())

	// 拿到响应数据之后获取trailer
	fmt.Printf("trailer:%#v\n", trailer)

	// respAdd, err := c.Add(ctx, &pb.AddParam{
	//   X: int32(*paramX),
	//   Y: int32(*paramY),
	// })
	// if err != nil {
	// 	log.Printf("c.Add failed, err:%v", err)
	// 	return
	// }

	// 调用服务端流式rpc
	// callLotsOfReplies(c)

	// callLotsOfGreetings(c)

	// runBidiHello(c)
}

func callLotsOfReplies(c pb.GreeterClient) {
	// server端流式RPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := c.LotsOfReqlies(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("c.LotsOfReplies failed, err: %v", err)
	}
	for {
		// 接收服务端返回的流式数据，当收到io.EOF或错误时退出
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("c.LotsOfReplies failed, err: %v", err)
		}
		log.Printf("got reply: %q\n", res.GetReply())
	}
}

func callLotsOfGreetings(c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 客户端要流式的发送请求消息
	stream, err := c.LotsOfGreetings(ctx)
	if err != nil {
		log.Printf("c.LotsOfGreetings(ctx) failed , err:%v\n", err)
		return
	}
	names := []string{"张三", "李四", "冯敏远"}
	for _, name := range names {
		stream.Send(&pb.HelloRequest{Name: name})
	}
	// 流式发送结束后要关闭流
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("stream.CloseAndRecv() failed , err:%v\n", err)
		return
	}
	log.Printf("res:%v\n", res.GetReply())
}

func runBidiHello(c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// 双向流模式
	stream, err := c.BidiHello(ctx)
	if err != nil {
		log.Fatalf("c.BidiHello failed, err: %v", err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			// 接收服务端返回的响应
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("c.BidiHello stream.Recv() failed, err: %v", err)
			}
			fmt.Printf("AI：%s\n", in.GetReply())
		}
	}()
	// 从标准输入获取用户输入
	reader := bufio.NewReader(os.Stdin) // 从标准输入生成读对象
	for {
		cmd, _ := reader.ReadString('\n') // 读到换行
		cmd = strings.TrimSpace(cmd)
		if len(cmd) == 0 {
			continue
		}
		if strings.ToUpper(cmd) == "QUIT" {
			break
		}
		// 将获取到的数据发送至服务端
		if err := stream.Send(&pb.HelloRequest{Name: cmd}); err != nil {
			log.Fatalf("c.BidiHello stream.Send(%v) failed: %v", cmd, err)
		}
	}
	stream.CloseSend()
	<-waitc
}
