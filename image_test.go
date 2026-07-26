package wordingo

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

// makeRedPNG returns a 2x2 red PNG with 72 DPI.
func makeRedPNG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic("encode red PNG: " + err.Error())
	}
	return buf.Bytes()
}

// makeSmallJPEG returns a 1x1 black JPEG
func makeSmallJPEG() []byte {
	img := image.NewGray(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic("encode JPEG: " + err.Error())
	}
	return buf.Bytes()
}

func TestImage_AddImageBytes(t *testing.T) {
	pngData := makeRedPNG()
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	// Add a paragraph first so there's a target paragraph for the image
	doc.AddParagraph("Before image")

	run, err := doc.AddImageBytes("test.png", pngData, "image/png")
	if err != nil {
		t.Fatalf("AddImageBytes() error: %v", err)
	}
	if run == nil {
		t.Fatal("AddImageBytes returned nil Run")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Verify media part exists in zip
	mediaContent := readZipEntryFromBuf(t, buf.Bytes(), "word/media/image1.png")
	if len(mediaContent) == 0 {
		t.Error("word/media/image1.png part is empty")
	}

	// Verify content type override
	ctContent := readZipEntryFromBuf(t, buf.Bytes(), "[Content_Types].xml")
	if !strings.Contains(ctContent, `image/png`) {
		t.Error("[Content_Types].xml missing image/png override")
	}
	if !strings.Contains(ctContent, `/word/media/image1.png`) {
		t.Error("[Content_Types].xml missing /word/media/image1.png override path")
	}

	// Verify relationship from document.xml.rels
	relsContent := readZipEntryFromBuf(t, buf.Bytes(), "word/_rels/document.xml.rels")
	if !strings.Contains(relsContent, relImage) {
		t.Error("document.xml.rels missing image relationship type")
	}
	if !strings.Contains(relsContent, "media/image1.png") {
		t.Error("document.xml.rels missing image target")
	}

	// Verify DrawingML inline element in document.xml
	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")
	if !strings.Contains(docContent, "<w:drawing") {
		t.Error("document.xml missing w:drawing element")
	}
	if !strings.Contains(docContent, "wp:inline") {
		t.Error("document.xml missing wp:inline element")
	}
	if !strings.Contains(docContent, "a:blip") {
		t.Error("document.xml missing a:blip element")
	}
	// pic element uses inline namespace (no pic: prefix in serialized output)
	if !strings.Contains(docContent, "xmlns=\"http://schemas.openxmlformats.org/drawingml/2006/picture\"") {
		t.Error("document.xml missing pic namespace")
	}
	if !strings.Contains(docContent, "<pic ") {
		t.Error("document.xml missing pic element")
	}
}

func TestImage_SetImageWidthHeight(t *testing.T) {
	pngData := makeRedPNG()
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.AddParagraph("")
	run, err := doc.AddImageBytes("test.png", pngData, "image/png")
	if err != nil {
		t.Fatalf("AddImageBytes() error: %v", err)
	}

	// Set width and height explicitly
	run.SetImageWidth(2.0).SetImageHeight(1.5)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	docContent := readZipEntryFromBuf(t, buf.Bytes(), "word/document.xml")

	// 2 inches = 2 * 914400 = 1828800 EMU
	expectedCX := `cx="1828800"`
	// 1.5 inches = 1.5 * 914400 = 1371600 EMU
	expectedCY := `cy="1371600"`

	if !strings.Contains(docContent, expectedCX) {
		t.Errorf("document.xml missing inline extent cx=1828800 (2in), got content:\n%s", docContent)
	}
	if !strings.Contains(docContent, expectedCY) {
		t.Errorf("document.xml missing inline extent cy=1371600 (1.5in), got content:\n%s", docContent)
	}
}

func TestImage_AddImageBytes_JPEG(t *testing.T) {
	jpegData := makeSmallJPEG()
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.AddParagraph("")
	run, err := doc.AddImageBytes("test.jpg", jpegData, "image/jpeg")
	if err != nil {
		t.Fatalf("AddImageBytes JPEG error: %v", err)
	}
	if run == nil {
		t.Fatal("AddImageBytes returned nil")
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Verify media part exists (extension should be .jpg)
	readZipEntryFromBuf(t, buf.Bytes(), "word/media/image1.jpg")

	ctContent := readZipEntryFromBuf(t, buf.Bytes(), "[Content_Types].xml")
	if !strings.Contains(ctContent, `image/jpeg`) {
		t.Error("[Content_Types].xml missing image/jpeg override")
	}
}

func TestImage_AddImageBytes_EmptyData(t *testing.T) {
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	doc.AddParagraph("")
	_, err = doc.AddImageBytes("empty.png", []byte{}, "image/png")
	if err == nil {
		t.Error("AddImageBytes with empty data should return error")
	}
}

func TestImage_MultipleImages(t *testing.T) {
	pngData := makeRedPNG()
	doc, err := Create()
	if err != nil {
		t.Fatal(err)
	}

	doc.AddParagraph("")

	_, err = doc.AddImageBytes("img1.png", pngData, "image/png")
	if err != nil {
		t.Fatalf("first AddImageBytes: %v", err)
	}
	_, err = doc.AddImageBytes("img2.png", pngData, "image/png")
	if err != nil {
		t.Fatalf("second AddImageBytes: %v", err)
	}

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// Both media parts should exist
	readZipEntryFromBuf(t, buf.Bytes(), "word/media/image1.png")
	readZipEntryFromBuf(t, buf.Bytes(), "word/media/image2.png")
}

func TestImage_DPI_Detection_PNG_72DPI(t *testing.T) {
	// Default PNG without pHYs chunk should return 72 DPI
	data := makeRedPNG()
	dpi := detectImageDPI(data)
	// The synthetic PNG has no pHYs, so should default to 72
	if dpi != 72 {
		t.Errorf("detectImageDPI = %d, want 72 (no pHYs chunk)", dpi)
	}
}

func TestImage_DPI_Detection_JPEG(t *testing.T) {
	data := makeSmallJPEG()
	dpi := detectImageDPI(data)
	// Small synthetic JPEG may or may not have JFIF marker.
	// Just verify it doesn't panic and returns something reasonable.
	if dpi <= 0 {
		t.Errorf("detectImageDPI JPEG = %d, want > 0", dpi)
	}
}

func TestImage_contentTypeDetection(t *testing.T) {
	pngData := makeRedPNG()
	jpegData := makeSmallJPEG()

	if ct := detectContentType(pngData); ct != "image/png" {
		t.Errorf("detectContentType(png) = %q, want image/png", ct)
	}
	if ct := detectContentType(jpegData); ct != "image/jpeg" {
		t.Errorf("detectContentType(jpeg) = %q, want image/jpeg", ct)
	}
	if ct := detectContentType([]byte{0, 0, 0}); ct != "" {
		t.Errorf("detectContentType(unknown) = %q, want empty", ct)
	}
}
