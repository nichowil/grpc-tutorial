package main

import (
	"context"
	"log"
	"time"

	pb "nichowil/grpc-tutorial/transform"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTransformClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SimulateError(ctx, &pb.ErrorHandlingRequest{Message: "invalid argument"})
	if err != nil {
		log.Printf("could not greet: %v\n", err) // log.Fatal stop apps when called
	}
	log.Printf("Greeting: %s", r.GetMessage())

	r, err = c.SimulateError(ctx, &pb.ErrorHandlingRequest{Message: "timeout"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
