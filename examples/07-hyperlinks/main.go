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

	doc.AddParagraph("Hyperlinks in Wordingo").SetStyle("Title")

	doc.AddParagraph("Text with Inline Hyperlinks").SetStyle("Heading2")

	p := doc.AddParagraph("Visit the ")
	p.AddHyperlink("Go Programming Language", "https://go.dev").SetColor("0563C1").SetUnderline("single")
	p.AddRun(" website to download the latest version.")

	doc.AddParagraph("")

	doc2 := doc.AddParagraph("For documentation, see ")
	doc2.AddHyperlink("pkg.go.dev", "https://pkg.go.dev").SetBold(true).SetColor("0563C1")

	doc.AddParagraph("")

	doc.AddParagraph("Multiple Links in One Paragraph").SetStyle("Heading2")

	para := doc.AddParagraph("Check out ")
	para.AddHyperlink("Google", "https://google.com").SetColor("2E75B6")
	para.AddRun(", ")
	para.AddHyperlink("GitHub", "https://github.com").SetColor("2E75B6")
	para.AddRun(", and ")
	para.AddHyperlink("Stack Overflow", "https://stackoverflow.com").SetColor("2E75B6")
	para.AddRun(" for more resources.")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
}
