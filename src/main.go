package main

import (
	"image"
	"image/color"
	"image/jpeg" // Register JPEG format
	"image/png"  // Register PNG  format
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	redLightnessContribution   = 0.21
	greenLightnessContribution = 0.72
	blueLightnessContribution  = 0.07
)

func lightness(color color.Color) float64 {
	r, g, b, a := color.RGBA()
	alphaRescaleFactor := float64(a)
	redContribution := (redLightnessContribution * float64(r)) / alphaRescaleFactor
	greenContribution := (greenLightnessContribution * float64(g)) / alphaRescaleFactor
	blueContribution := (blueLightnessContribution * float64(b)) / alphaRescaleFactor
	return redContribution + greenContribution + blueContribution
}

func drawCircle(img image.Paletted, center image.Point, radius int, fillColor color.Color) {
	for x := center.X - radius; x < center.X+radius; x++ {
		for y := center.Y - radius; y < center.Y+radius; y++ {
			currentRadius := math.Pow(float64(x-center.X), 2) + math.Pow(float64(y-center.Y), 2)
			if currentRadius < float64(radius) {
				img.Set(x, y, fillColor)
			}
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		log.Fatalln("Needs three arguments")
	}

	inFile, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
	defer inFile.Close()

	img, _, err := image.Decode(inFile)
	if err != nil {
		log.Fatalln(err)
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	fidelity, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatalln(err)
	}

	kernelSize := int(math.Floor(math.Min(float64(width), float64(height)) / float64(fidelity)))
	halfKernelSize := kernelSize / 2

	blotchImg := image.NewPaletted(bounds, color.Palette{
		color.White,
		color.Black,
	})

	for x := halfKernelSize; x < width-halfKernelSize; x += kernelSize {
		for y := halfKernelSize; y < height-halfKernelSize; y += kernelSize {
			weight := 0.0
			count := 0.0
			for kernel_x := -halfKernelSize; kernel_x < halfKernelSize; kernel_x += 1 {
				for kernel_y := -halfKernelSize; kernel_y < halfKernelSize; kernel_y += 1 {
					weight += lightness(img.At(x+kernel_x, y+kernel_y))
					count += 1
				}
			}
			blotchScalar := 1 - weight/count
			radius := int(float64(kernelSize) * blotchScalar)
			if radius > 0 {
				drawCircle(*blotchImg, image.Point{x, y}, radius, blotchImg.Palette[1])
			}
		}
	}

	outFile, err := os.Create(os.Args[3])
	if err != nil {
		log.Fatalln(err)
	}
	defer outFile.Close()

	if strings.HasSuffix(outFile.Name(), ".png") {
		png.Encode(outFile, blotchImg)
	} else if strings.HasSuffix(outFile.Name(), ".jpg") || strings.HasSuffix(outFile.Name(), ".jpeg") {
		jpeg.Encode(outFile, blotchImg, &jpeg.Options{Quality: 100})
	} else {
		panic("Expected output file to be of type .png or .jpeg")
	}
}
