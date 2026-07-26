package main

import (
	"fmt"
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

	doc.AddHeader(wordingo.HeaderDefault).AddParagraph("Report — Confidential")
	doc.AddFooter(wordingo.FooterDefault).AddParagraph("Page ")

	doc.SetPaperSize(wordingo.PaperA4W, wordingo.PaperA4H)
	doc.SetMargins(1440, 1440, 1440, 1440)

	doc.SetOrientation(wordingo.OrientationLandscape)

	doc.AddParagraph("Page Setup Example").SetStyle("Title")

	doc.AddParagraph("This document uses A4 paper size, landscape orientation, and 1-inch margins. A header and footer are active on every page.")

	doc.AddParagraph("").SetPageBreakBefore(true)
	doc.AddParagraph("Page 2 Content").SetStyle("Heading1")
	doc.AddParagraph("This paragraph starts on a new page via SetPageBreakBefore.")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
}
