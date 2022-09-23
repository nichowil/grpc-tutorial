package main

import (
	"flag"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
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
	flag.Parse()
	var (
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

	waitc := make(chan struct{})
	go func() {
		close(waitc)
	}()

	<-waitc

	newImage := vectorToImage(newImageV)
	err = saveImageToFilePath(newImage, "images/result.jpg")
	if err != nil {
		log.Fatalf("Failed to save image : %v", err)
	}

}
