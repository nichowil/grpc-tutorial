package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

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
		time.Sleep(time.Second * 5)
	}

	return &pb.ErrorHandlingResponse{Message: "Testing error code : " + in.GetMessage()}, nil
}

// SayHello implements helloworld.TransformServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloResponse{Message: "Hello " + in.GetName()}, nil
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
	s := grpc.NewServer(grpc.StreamInterceptor(StreamServerInterceptor), grpc.UnaryInterceptor(UnaryServerTimeoutInterceptor))
	pb.RegisterTransformServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("[Intercept request] : pre request interceptor")

	err := handler(srv, ss)
	if err != nil {
		return err
	}

	log.Println("[Intercept request] : post request interceptor")
	return nil
}

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Invoke 'handler' to use your gRPC server implementation and get
	// the response.
	log.Println("[Intercept request] : pre request interceptor")

	// Get the metadata from the incoming context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("couldn't parse incoming context metadata")
	}
	log.Println("metadata : ", md)

	h, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}

	log.Println("[Intercept request] : post request interceptor")
	return h, err
}

func UnaryServerTimeoutInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var err error
	var result interface{}

	log.Println("[Intercept request] : pre request interceptor")

	done := make(chan struct{})

	go func() {
		result, err = handler(ctx, req)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Println("context timed out")
			return nil, status.New(codes.Canceled, "Client cancelled, abandoning.").Err()
		}
	case <-done:
	}

	log.Println("[Intercept request] : post request interceptor")
	return result, err
}
