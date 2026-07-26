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

	doc.AddParagraph("The Art of Go").SetStyle("Title")

	doc.AddParagraph("Simple, Fast, Reliable").SetStyle("Subtitle")

	doc.AddParagraph("Chapter 1: Getting Started").SetStyle("Heading1")

	doc.AddParagraph("Go is a statically typed, compiled programming language designed at Google. It is known for its simplicity, strong concurrency primitives, and fast build times.")

	doc.AddParagraph("Why Go?").SetStyle("Heading2")

	p := doc.AddParagraph("Go offers a clean syntax, built-in concurrency via goroutines, and a rich standard library.")
	p.AddRun(" This sentence is bold.").SetBold(true)
	p.AddRun(" This one is italic.").SetItalic(true)

	doc.AddParagraph("Performance").SetStyle("Heading3")

	r := doc.AddParagraph("")
	r.AddRun("Compiled to native code — ").SetColor("666666")
	r.AddRun("lightning fast execution").SetBold(true).SetColor("2E75B6")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
}
