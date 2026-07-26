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

	doc.AddHeader(wordingo.HeaderDefault).AddParagraph("Wordingo Demo Document — Header")
	doc.AddFooter(wordingo.FooterDefault).AddParagraph("Page ")

	doc.SetPaperSize(wordingo.PaperLetterW, wordingo.PaperLetterH)

	doc.AddParagraph("Wordingo: Pure Go Word Documents").SetStyle("Title")

	doc.AddParagraph("A comprehensive demonstration of every library feature").SetStyle("Subtitle")

	doc.AddParagraph("1. Text & Styles").SetStyle("Heading1")

	doc.AddParagraph("Wordingo is a zero-dependency Go library for creating and editing WordprocessingML (.docx) documents. It supports styled paragraphs, tables, images, headers, footers, lists, hyperlinks, and page setup — all in pure Go.")

	p := doc.AddParagraph("Inline formatting: ")
	p.AddRun("bold").SetBold(true)
	p.AddRun(", ")
	p.AddRun("italic").SetItalic(true)
	p.AddRun(", ")
	p.AddRun("underlined").SetUnderline("single")
	p.AddRun(", ")
	p.AddRun("colored text").SetColor("E74C3C")
	p.AddRun(", and ")
	p.AddRun("styled runs").SetFont("Courier New").SetSize(10)

	doc.AddParagraph("2. Tables").SetStyle("Heading1")

	doc.AddParagraph("Simple Grid Table").SetStyle("Heading2")

	doc.AddTable([][]string{
		{"Feature", "Status", "Version"},
		{"Paragraphs", "Done", "1.0"},
		{"Tables", "Done", "1.0"},
		{"Images", "Done", "1.0"},
		{"Headers/Footers", "Done", "1.0"},
		{"Lists", "Done", "1.0"},
		{"Hyperlinks", "Done", "1.0"},
	})

	doc.AddParagraph("Complex Builder Table").SetStyle("Heading2")

	tbl := doc.AddTableBuilder()
	tbl.SetTableStyle("LightGridAccent1")
	tbl.SetWidth(8000, "dxa")
	tbl.SetBorders(&wordingo.TableBorders{
		Top:    &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
		Bottom: &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
	})

	tbl.Row(0).Cell(0).SetText("Item").SetBold(true).SetWidth(2500, "dxa").SetShading("", "D9E2F3")
	tbl.Row(0).Cell(1).SetText("Description").SetBold(true).SetWidth(3500, "dxa").SetShading("", "D9E2F3")
	tbl.Row(0).Cell(2).SetText("Price").SetBold(true).SetWidth(1000, "dxa").SetShading("", "D9E2F3")

	tbl.Row(1).Cell(0).SetText("Widget")
	tbl.Row(1).Cell(1).SetText("A useful widget with many features")
	tbl.Row(1).Cell(2).SetText("$19.99")

	tbl.Row(2).Cell(0).SetText("Gadget")
	tbl.Row(2).Cell(1).SetText("Portable and lightweight")
	tbl.Row(2).Cell(2).SetText("$24.99")

	tbl.Row(3).Cell(0).SetText("Toolkit")
	tbl.Row(3).Cell(1).SetText("Complete set of tools")
	tbl.Row(3).Cell(2).SetText("$49.99")

	doc.AddParagraph("3. Images").SetStyle("Heading1")

	imgData := makeGradientPNG()
	doc.AddParagraph("")
	run, err := doc.AddImageBytes("gradient.png", imgData, "image/png")
	if err != nil {
		log.Fatal(err)
	}
	run.SetImageWidth(2.5).SetImageHeight(2.5)

	doc.AddParagraph("A 100x100 gradient PNG embedded via DrawingML inline with locked aspect ratio.")

	doc.AddParagraph("4. Lists").SetStyle("Heading1")

	doc.AddParagraph("Key Features (Bulleted)").SetStyle("Heading2")

	features := doc.AddList(false)
	features.AddItem("Pure Go — zero external dependencies", 0)
	features.AddItem("Fluent builder API for tables and lists", 0)
	features.AddItem("DrawingML inline images with DPI detection", 0)
	features.AddItem("Header/footer support with OPC part creation", 0)
	features.AddItem("Page setup (orientation, paper size, margins)", 0)
	features.AddItem("External hyperlinks with Run chaining", 0)

	doc.AddParagraph("Implementation Plan (Ordered)").SetStyle("Heading2")

	plan := doc.AddList(true)
	plan.AddItem("Foundation: OPC package, styles, basic text", 0)
	plan.AddItem("Style Engine: clone, resolve, apply", 0)
	plan.AddItem("Document Model: paragraph, run, formatting", 0)
	plan.AddItem("Content API: tables, images, headers, lists", 0)
	plan.AddItem("Rich Content: numbering, hyperlinks, page setup", 0)

	doc.AddParagraph("5. Hyperlinks").SetStyle("Heading1")

	link := doc.AddParagraph("Learn more at ")
	link.AddHyperlink("GitHub Repository", "https://github.com/fabiomarini/wordingo").SetColor("0563C1").SetUnderline("single")

	doc.AddParagraph("6. Page Setup").SetStyle("Heading1")

	doc.AddParagraph("The first section of this document uses Letter size with portrait orientation and 1-inch margins. Page breaks and landscape sections are demonstrated below.")

	doc.AddParagraph("").SetPageBreakBefore(true)

	doc.SetOrientation(wordingo.OrientationLandscape)
	doc.SetMargins(720, 720, 720, 720)

	doc.AddParagraph("Landscape Section").SetStyle("Heading1")

	doc.AddParagraph("This section uses landscape orientation with 0.5-inch margins. All page setup methods (SetOrientation, SetPaperSize, SetMargins) are chainable on Document.")

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
