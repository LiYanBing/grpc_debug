package main

import (
	"context"
	"errors"
	"log"
	"net"

	"google.golang.org/grpc"

	example "github.com/LiYanBing/grpc_debug/example/api"
	_ "github.com/LiYanBing/grpc_debug/grpc_encoding/json"
)

func main() {
	listen, err := net.Listen("tcp", ":4096")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	example.RegisterExampleServiceServer(server, &Example{})
	err = server.Serve(listen)
	if err != nil {
		log.Fatal(err)
	}
}

type Example struct {
}

func (s *Example) GetName(ctx context.Context, args *example.GetNameRequest) (*example.GetNameResponse, error) {
	if args == nil {
		return nil, errors.New("empty args")
	}

	return &example.GetNameResponse{
		Name: args.Name,
		Age:  args.Age,
	}, nil
}
