package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	v1 "go-grpc/api/service/v1"
	"go-grpc/server"
)

type Config struct {
	GRPCPort   string
	DBHost     string
	DBUser     string
	DBPassword string
	DBSchema   string
}

func RunServer() error {
	ctx := context.Background()
	var cfg Config

	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "grpc-port")
	flag.StringVar(&cfg.DBHost, "db-host", "", "db-host")
	flag.StringVar(&cfg.DBUser, "db-user", "", "db-user")
	flag.StringVar(&cfg.DBPassword, "db-pwd", "", "db-pwd")
	flag.StringVar(&cfg.DBSchema, "db-schema", "", "db-schema")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port of gRPC: '%s'", cfg.GRPCPort)
	}

	param := "parseTime=true"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBSchema,
		param,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败 %v", err.Error())
	}
	defer db.Close()
	v1API := v1.NewToDoServiceServer(db)
	return server.RunServer(ctx, v1API, cfg.GRPCPort)
}
