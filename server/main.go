package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"net"

	pb "nichowil/grpc-tutorial/transform"

	"google.golang.org/grpc"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 50051, "The server port")
)

type imageVector struct {
	r float32
	g float32
	b float32
	a float32
}

func imageToVector(ctx context.Context, img image.Image) (res [][]imageVector) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	res = make([][]imageVector, height)
	for y := 0; y < height; y++ {
		res[y] = make([]imageVector, width)
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			res[y][x] = imageVector{
				r: float32(r),
				g: float32(g),
				b: float32(b),
				a: float32(a),
			}
		}
	}
	return
}

func vectorToImage(ctx context.Context, res [][]imageVector) (img image.Image) {
	height := len(res)
	width := len(res[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	tmpImg := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for y := 0; y < height; y++ {
		res[y] = make([]imageVector, width)
		for x := 0; x < width; x++ {
			pixel := color.RGBA{
				uint8(res[y][x].r),
				uint8(res[y][x].g),
				uint8(res[y][x].b),
				uint8(res[y][x].a),
			}
			tmpImg.Set(x, y, pixel)
		}
	}
	img = tmpImg
	return
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
