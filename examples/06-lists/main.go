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

	doc.AddParagraph("Lists in Wordingo").SetStyle("Title")

	doc.AddParagraph("Shopping List (Bulleted)").SetStyle("Heading2")

	list := doc.AddList(false)
	list.AddItem("Apples", 0)
	list.AddItem("Bananas", 0)
	list.AddItem("Milk", 0)
	list.AddItem("Bread", 0)

	doc.AddParagraph("")

	doc.AddParagraph("Recipe Steps (Ordered)").SetStyle("Heading2")

	steps := doc.AddList(true)
	steps.AddItem("Preheat oven to 350°F", 0)
	steps.AddItem("Mix flour and sugar", 0)
	steps.AddItem("Add eggs and vanilla", 0)
	steps.AddItem("Pour into baking pan", 0)
	steps.AddItem("Bake for 30 minutes", 0)

	doc.AddParagraph("")

	doc.AddParagraph("Nested Outline").SetStyle("Heading2")

	outline := doc.AddList(true)
	outline.AddItem("Project Setup", 0)
	outline.AddItem("Install dependencies", 1)
	outline.AddItem("Configure database", 1)
	outline.AddItem("Development", 0)
	outline.AddItem("Write backend API", 1)
	outline.AddItem("Validation", 2)
	outline.AddItem("Error handling", 2)
	outline.AddItem("Build frontend", 1)
	outline.AddItem("Deployment", 0)
	outline.AddItem("CI/CD pipeline", 1)
	outline.AddItem("Monitoring", 1)

	doc.AddParagraph("")

	doc.AddParagraph("Convenience (AddListFromSlice)").SetStyle("Heading2")

	doc.AddListFromSlice([]string{"One", "Two", "Three", "Four"}, true)

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
}
