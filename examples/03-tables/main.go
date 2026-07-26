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

	doc.AddParagraph("Tables in Wordingo").SetStyle("Title")

	doc.AddParagraph("Simple Grid Table").SetStyle("Heading2")

	doc.AddTable([][]string{
		{"Name", "Language", "Paradigm"},
		{"Go", "Statically typed", "Compiled"},
		{"Python", "Dynamically typed", "Interpreted"},
		{"Rust", "Statically typed", "Compiled"},
	})

	doc.AddParagraph("")

	doc.AddParagraph("Complex Builder Table").SetStyle("Heading2")

	tbl := doc.AddTableBuilder()
	tbl.SetTableStyle("LightGridAccent1")
	tbl.SetBorders(&wordingo.TableBorders{
		Top:    &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
		Bottom: &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
	})

	tbl.Row(0).Cell(0).SetText("Product").SetBold(true)
	tbl.Row(0).Cell(1).SetText("Price").SetBold(true).SetWidth(1500, "dxa")
	tbl.Row(0).Cell(2).SetText("Rating").SetBold(true).SetWidth(1200, "dxa")

	tbl.Row(1).Cell(0).SetText("Widget Pro")
	tbl.Row(1).Cell(1).SetText("$29.99")
	tbl.Row(1).Cell(2).SetText("4.8")

	tbl.Row(2).Cell(0).SetText("Gadget Lite")
	tbl.Row(2).Cell(1).SetText("$14.99")
	tbl.Row(2).Cell(2).SetText("4.5")

	tbl.Row(3).Cell(0).SetText("Super Tool")
	tbl.Row(3).Cell(1).SetText("$49.99")
	tbl.Row(3).Cell(2).SetText("4.9")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
}
