package main

import (
	"context"
	"flag"
	"hello_client/pb"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpc 客户端
// 调用server端的 SayHello 方法

var name = flag.String("name", "tang", "通过-name告诉server你是谁")
var paramX = flag.Int("paramX" , 1, "通过-paramX输入paramX")
var paramY = flag.Int("paramY" , 2, "通过-paramY输入paramY")

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
	// 调用RPC方法
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	respHello, err1 := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
  respAdd, err2 := c.Add(ctx, &pb.AddParam{
    X: int32(*paramX),
    Y: int32(*paramY),
  })
	if err1 != nil || err2 != nil {
		log.Printf("c.SayHello or c.Add failed, err:%v", err)
		return
	}
	// 拿到了RPC响应
	log.Printf("resp:%v", respHello.GetReply())
  log.Printf("resp:%v", respAdd.GetZ())
}
