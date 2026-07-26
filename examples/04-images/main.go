package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/fabiomarini/wordingo"
)

func main() {
	doc, err := wordingo.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	doc.AddParagraph("Images in Wordingo").SetStyle("Title")

	doc.AddParagraph("Embedded Images").SetStyle("Heading2")

	imgData := makeGradientPNG()
	doc.AddParagraph("")
	run, err := doc.AddImageBytes("gradient.png", imgData, "image/png")
	if err != nil {
		log.Fatal(err)
	}
	run.SetImageWidth(3.0).SetImageHeight(2.0)

	doc.AddParagraph("The image above is a 100x100 gradient PNG embedded as a DrawingML inline element with a locked aspect ratio and stretch-to-fill behavior.")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
}

func makeGradientPNG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			r := uint8(float64(x) / 100 * 255)
			b := uint8(float64(y) / 100 * 255)
			img.Set(x, y, color.NRGBA{R: r, G: 100, B: b, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
