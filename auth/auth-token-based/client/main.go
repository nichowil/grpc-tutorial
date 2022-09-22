package main

// https://chowdera.com/2022/199/202207181303421208.html CERT
// https://jbrandhorst.com/post/grpc-auth/ auth
// https://dev.to/techschoolguru/how-to-secure-grpc-connection-with-ssl-tls-in-go-4ph
// http://www.inanzzz.com/index.php/post/cvjx/using-oauth-authentication-tokens-for-grpc-client-and-server-communications-in-golang

import (
	"context"
	"log"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/credentials/oauth"

	pb "nichowil/grpc-tutorial/transform"
)

func main() {

	rpcCreds := oauth.NewOauthAccess(&oauth2.Token{AccessToken: "contoh_token"})

	creds, err := credentials.NewClientTLSFromFile("./auth/cert/ca-cert.pem", "localhost")
	if err != nil {
		log.Fatalf("error to load TLS : %+v", err)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(rpcCreds),
	}
	// opts = append(opts, grpc.WithBlock())

	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTransformClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "Huda testing"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
