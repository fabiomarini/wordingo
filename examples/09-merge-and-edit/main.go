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

	doc.AddParagraph("Template Merge & Edit Demo").SetStyle("Title")
	doc.AddParagraph("This example demonstrates Phase 6 features: template merge ({{placeholder}} replacement) and edit operations (insert, delete, replace).")

	h := doc.AddHeader(wordingo.HeaderDefault)
	h.AddParagraph("Report: {{report_name}} — {{date}}")

	f := doc.AddFooter(wordingo.FooterDefault)
	f.AddParagraph("Page {{page_num}}")

	doc.AddParagraph("1. Merge: Body Placeholders").SetStyle("Heading1")
	doc.AddParagraph("Hello {{name}}, your {{item}} is ready for {{action}}.")
	doc.AddParagraph("Reference: {{ref_id}}")

	doc.AddParagraph("2. Merge: Table Cell Placeholders").SetStyle("Heading1")
	doc.AddTable([][]string{
		{"Product", "Qty", "Price", "Total"},
		{"{{prod_1}}", "{{qty_1}}", "{{price_1}}", "{{total_1}}"},
		{"{{prod_2}}", "{{qty_2}}", "{{price_2}}", "{{total_2}}"},
	})

	doc.Merge(map[string]string{
		"report_name": "Q2 Sales Report",
		"date":        "2026-07-26",
		"page_num":    "1",
		"name":        "Alice",
		"item":        "invoice",
		"action":      "review",
		"ref_id":      "INV-2026-0420",
		"prod_1":      "Widget",   "qty_1": "10", "price_1": "$19.99", "total_1": "$199.90",
		"prod_2":      "Gadget",   "qty_2": "5",  "price_2": "$24.99", "total_2": "$124.95",
	}, nil)

	if warnings := doc.Warnings(); len(warnings) > 0 {
		for _, w := range warnings {
			fmt.Printf("Merge warning: %s\n", w)
		}
	}

	doc.AddParagraph("3. Edit: InsertBefore / InsertAfter").SetStyle("Heading1")

	paras := doc.Paragraphs()
	var introPara *wordingo.Paragraph
	for _, p := range paras {
		if p.Text() == "Hello Alice, your invoice is ready for review." {
			introPara = p
			break
		}
	}
	if introPara != nil {
		doc.InsertBefore(introPara, "Dear Alice,")
		doc.InsertAfter(introPara, "Please find the details below.")
	}

	doc.AddParagraph("4. Edit: ReplaceText and SetText").SetStyle("Heading1")

	for _, p := range doc.Paragraphs() {
		if p.Text() == "Reference: INV-2026-0420" {
			p.AddRun(" (updated)").SetText(" (revised)")
		}
		if p.Text() == "Dear Alice," {
			p.AddRun(" Welcome!").SetText(" Welcome to the team!")
		}
	}

	doc.AddParagraph("5. Edit: DeleteParagraph").SetStyle("Heading1")

	var toDelete *wordingo.Paragraph
	for _, p := range doc.Paragraphs() {
		if p.Text() == "Page 1" {
			toDelete = p
			break
		}
	}
	if toDelete != nil {
		doc.DeleteParagraph(toDelete)
	}

	doc.AddParagraph("6. Body() Accessor — Document Order").SetStyle("Heading1")

	parts := make(map[string]int)
	for _, el := range doc.Body() {
		switch el.Type {
		case wordingo.ElementParagraph:
			parts["paragraphs"]++
		case wordingo.ElementTable:
			parts["tables"]++
		}
	}
	doc.AddParagraph(fmt.Sprintf("Body contains %d paragraphs and %d tables in document order.", parts["paragraphs"], parts["tables"]))

	doc.AddParagraph("7. Edit: DeleteRow").SetStyle("Heading1")

	tables := doc.Tables()
	if len(tables) > 0 {
		if err := tables[0].DeleteRow(2); err != nil {
			log.Fatalf("DeleteRow: %v", err)
		}
		doc.AddParagraph("Row 2 (Gadget) deleted from table.")
	}

	doc.AddParagraph("8. Scoped Merge — Headers Only").SetStyle("Heading1")
	doc.Merge(map[string]string{"page_num": "2"}, &wordingo.MergeOpts{
		ScopedParts: wordingo.ScopedParts{Headers: true, Footers: true},
	})

	doc.AddParagraph("All Phase 6 features demonstrated. The final document has been edited through insertions, deletions, text replacements, and scoped merge operations — all verified through round-trip safety.")

	doc.AddParagraph("Created with wordingo v0.1.0 — Merge, Edit, BodyElement, InsertBefore/After, DeleteParagraph, DeleteRow, ReplaceText, SetText.").SetStyle("Subtitle")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}

	fi, _ := os.Stat(out)
	fmt.Printf("Created %s (%d bytes)\n", out, fi.Size())
	warnings := doc.Warnings()
	if len(warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range warnings {
			fmt.Printf("  %s\n", w)
		}
	}
	fmt.Println("\nPhase 6 features demonstrated:")
	fmt.Println("  - Merge() with body, table, header, footer placeholder replacement")
	fmt.Println("  - MergeOpts.ScopedParts (headers/footers only merge)")
	fmt.Println("  - InsertBefore / InsertAfter by pointer identity")
	fmt.Println("  - DeleteParagraph by pointer identity")
	fmt.Println("  - DeleteRow with bounds safety")
	fmt.Println("  - Run.SetText and Run.ReplaceText")
	fmt.Println("  - Body() accessor for document-order iteration")
	fmt.Println("  - Warnings() for merge and edit operations")
}
