package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "nichowil/grpc-tutorial/transform"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := c.SimulateError(ctx, &pb.ErrorHandlingRequest{Message: "invalid argument"})
	if err != nil {
		log.Printf("could not simulate error: %v\n", err) // log.Fatal stop apps when called
	}

	r, err = c.SimulateError(ctx, &pb.ErrorHandlingRequest{Message: "timeout"})
	if err != nil {
		log.Printf("could not simulate error: %v", err)
	}

	r, err = c.SimulateError(ctx, &pb.ErrorHandlingRequest{Message: "detail"})
	if err != nil {
		log.Printf("could not simulate error: %v", err)
		st := status.Convert(err)
		for _, detail := range st.Details() {
			switch t := detail.(type) {
			case *errdetails.BadRequest:
				fmt.Println("Oops! Your request was rejected by the server.")
				for _, violation := range t.GetFieldViolations() {
					fmt.Printf("The %q field was wrong:\n", violation.GetField())
					fmt.Printf("\t%s\n", violation.GetDescription())
				}
			}
		}
	}

	log.Printf("Response: %s", r.GetMessage())
}
