package server

import (
	"context"
	v1 "go-grpc/api/proto/v1"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func RunServer(ctx context.Context, v1API v1.ToDoServiceServer, port string) error {
	listen, err := net.Listen("TCP", ":"+port)
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	v1.RegisterToDoServiceServer(server, v1API)

	channel := make(chan os.Signal, 1)
	go func() {
		for range channel {
			log.Println("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()
	log.Println("start gRPC server...")
	return server.Serve(listen)
}
