package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

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
	}

	return &pb.ErrorHandlingResponse{Message: "Testing error code : " + in.GetMessage()}, nil
}

func (s *server) Transform(stream pb.Transform_TransformServer) error {
	//var imageVector [][]pb.Color
	for {
		pixel, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// do smth to pixel here
		pixel.Color.R = 0

		stream.Send(pixel)
	}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.StreamInterceptor(StreamServerInterceptor))
	pb.RegisterTransformServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("Stream server Interceptor")
	return nil
}
