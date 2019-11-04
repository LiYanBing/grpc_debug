package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	example "github.com/LiYanBing/grpc_debug/example/api"
)

func main() {
	cert, err := credentials.NewClientTLSFromFile("./../conf/server.pem", "liulishuo.com")
	if err != nil {
		log.Fatalf("credentials.NewClientTLSFromFile err: %v", err)
	}

	conn, err := grpc.Dial(":4096", grpc.WithTransportCredentials(cert))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := example.NewExampleServiceClient(conn)
	ret, err := client.GetName(context.Background(), &example.GetNameRequest{
		Name: "liyanbing",
		Age:  18,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("result:%#v", *ret)
}
