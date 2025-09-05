package main

import (
	"dlmsprocessor/api"
	"dlmsprocessor/proto"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterDLMSProcessorServer(grpcServer, api.NewDLMSProcessorAPI())
	fmt.Println("Server is running on port 50051")
	grpcServer.Serve(lis)
}
