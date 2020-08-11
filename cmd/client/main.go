package client

import (
	"context"
	"flag"
	v1 "go-grpc/api/proto/v1"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"

	"google.golang.org/grpc"
)

const (
	appVersion = "v1"
)

func main() {
	address := flag.String("server", "", "gRPC server in format 'host:port'")
	flag.Parse()

	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatal("连接服务器失败...", err.Error())
	}
	defer conn.Close()

	client := v1.NewToDoServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t := time.Now().In(time.UTC)
	reminder, err := ptypes.TimestampProto(t)
	if err != nil {
		log.Fatal("ptypes.TimestampProto 失败")
	}

	pfx := t.Format(time.RFC3339Nano)

	// Create
	createRequest := &v1.CreateRequest{
		Api: appVersion,
		ToDo: &v1.ToDo{
			Title:       "title (" + pfx + ")",
			Description: "description(" + pfx + ")",
			Reminder:    reminder,
		},
	}
	res, err := client.Create(ctx, createRequest)
	if err != nil {
		log.Fatalf("创建失败 %v", err)
	}
	log.Printf("Create result -> %v\n", res)
}
