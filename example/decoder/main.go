package main

import (
	"flag"
	"image"
	"log"
	"os"

	"github.com/pixiv/go-libjpeg/jpeg"
)

func main() {
	flag.Parse()
	file := flag.Arg(0)

	io, err := os.Open(file)
	if err != nil {
		log.Fatalln("Can't open file: ", file)
	}

	img, err := jpeg.Decode(io, &jpeg.DecoderOptions{})
	if img == nil {
		log.Fatalln("Got nil")
	}
	if err != nil {
		log.Fatalln("Got Error: %v", err)
	}

	//
	// write your code here ...
	//

	switch img.(type) {
	case *image.YCbCr:
		log.Println("decoded YCbCr")
	case *image.Gray:
		log.Println("decoded Gray")
	default:
		log.Println("unknown format")
	}
}
