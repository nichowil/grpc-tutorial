package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	pb "nichowil/grpc-tutorial/transform"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type transformServer struct {
	pb.UnimplementedTransformServer
}

func (s *transformServer) Transform(stream pb.Transform_TransformServer) error {
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
	s := grpc.NewServer()
	pb.RegisterTransformServer(s, &transformServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
