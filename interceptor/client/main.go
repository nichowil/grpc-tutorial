package main

import (
	"context"
	"flag"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "nichowil/grpc-tutorial/transform"
)

var (
	imagePath  = flag.String("img", "images/test.jpg", "Image filepath")
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

type imageVector struct {
	r float32
	g float32
	b float32
	a float32
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}

func saveImageToFilePath(img image.Image, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err = jpeg.Encode(f, img, nil); err != nil {
		return err
	}
	return nil
}

func imageToVector(img image.Image) (res [][]imageVector) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	res = make([][]imageVector, height)
	for y := 0; y < height; y++ {
		res[y] = make([]imageVector, width)
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			res[y][x] = imageVector{
				r: float32(r >> 8),
				g: float32(g >> 8),
				b: float32(b >> 8),
				a: float32(a >> 8),
			}
		}
	}
	return
}

func vectorToImage(res [][]imageVector) (img image.Image) {
	height := len(res)
	width := len(res[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	tmpImg := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for y := 0; y < height; y++ {
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
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(SimpleLogInterceptor), grpc.WithStreamInterceptor(StreamClientInterceptor))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTransformClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = c.SimulateError(ctx, &pb.ErrorHandlingRequest{Message: "invalid argument"})
	if err != nil {
		log.Printf("could not simulate error: %v\n", err) // log.Fatal stop apps when called
	}

	_, err = c.SimulateError(ctx, &pb.ErrorHandlingRequest{Message: "timeout"})
	if err != nil {
		log.Printf("could not simulate error: %v", err)
	}

	rsuccess, err := c.SimulateError(context.Background(), &pb.ErrorHandlingRequest{Message: "success"})
	if err != nil {
		log.Printf("could not simulate error: %v\n", err)
	}

	log.Printf("Response: %s", rsuccess.GetMessage())

	flag.Parse()
	var (
		//opts      []grpc.DialOption
		newImageV [][]imageVector
	)

	image, err := getImageFromFilePath(*imagePath)
	if err != nil {
		log.Fatalf("fail to get image: %v", err)
	}

	bounds := image.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	newImageV = make([][]imageVector, height)
	for y := 0; y < height; y++ {
		newImageV[y] = make([]imageVector, width)
	}

	stream, err := c.Transform(context.Background())

	waitc := make(chan struct{})

	go func() {
		for {
			// receive next pixel from stream
			r, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}

			// update pixel on image vector
			newImageV[r.Point.Y][r.Point.X] = imageVector{
				r: r.Color.R,
				g: r.Color.G,
				b: r.Color.B,
				a: r.Color.A,
			}
		}
	}()

	imageV := imageToVector(image)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sendPixel := pb.Pixel{
				Color: &pb.Color{
					R: imageV[y][x].r,
					G: imageV[y][x].g,
					B: imageV[y][x].b,
					A: imageV[y][x].a,
				},
				Point: &pb.Point{
					X: int32(x),
					Y: int32(y),
				},
			}
			stream.Send(&sendPixel)
		}
	}

	stream.CloseSend()
	<-waitc

	newImage := vectorToImage(newImageV)
	saveImageToFilePath(newImage, "images/result.jpg")
}

func SimpleLogInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Println("Interceptor client : ", req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// Call 'streamer' to write messages to the stream before this function returns
	log.Println("Interceptor stream client")
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}

	return s, err
}
