package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	example "github.com/LiYanBing/grpc_debug/example/api"
)

func main() {
	conn, err := grpc.Dial(":4096", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

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
