package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "nichowil/grpc-tutorial/transform"

	"google.golang.org/grpc/status"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	pb.UnimplementedTransformServer
}

func (s *server) SimulateError(ctx context.Context, in *pb.ErrorHandlingRequest) (*pb.ErrorHandlingResponse, error) {
	log.Printf("Received: %v", in.GetMessage())

	if in.GetMessage() == "invalid argument" {
		log.Println("invalid argument : called")
		return &pb.ErrorHandlingResponse{}, status.Error(codes.InvalidArgument, "Max num of characters exceed")
	} else if in.GetMessage() == "timeout" {
		time.Sleep(time.Second * 2)
	}

	return &pb.ErrorHandlingResponse{Message: "Testing error code : " + in.GetMessage()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTransformServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
