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

	doc.AddParagraph("Hello, World!")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
}
