package main

import (
	"fmt"
	"log"

	"github.com/fabiomarini/wordingo"
)

func main() {
	// ============================================================
	// Part 1: Build a document with headings
	// ============================================================
	fmt.Println("=== Part 1: Build the document ===")

	doc, err := wordingo.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	doc.AddParagraph("Annual Report").SetStyle("Title")
	doc.AddParagraph("Fiscal Year 2026").SetStyle("Subtitle")

	doc.AddParagraph("Executive Summary").SetStyle("Heading1")
	doc.AddParagraph("Revenue grew 12% year over year. Key initiatives shipped on time.")
	doc.AddParagraph("Highlights").SetStyle("Heading2")
	doc.AddParagraph("Three product lines exceeded their targets.")
	doc.AddParagraph("Regional Performance").SetStyle("Heading2")
	doc.AddParagraph("Europe led growth, APAC followed closely.")
	doc.AddParagraph("Financials").SetStyle("Heading1")
	doc.AddParagraph("A summary of the income statement follows.")
	doc.AddParagraph("Outlook").SetStyle("Heading1")
	doc.AddParagraph("We expect continued momentum into next year.")

	// ============================================================
	// Part 2: Inspect the outline (Headings)
	// ============================================================
	fmt.Println("\n=== Part 2: Headings() ===")

	for _, h := range doc.Headings() {
		fmt.Printf("  L%d  %-22s [%s]\n", h.Level, h.Text, h.Style)
	}

	// ============================================================
	// Part 3: Add a table of contents at the start
	// ============================================================
	fmt.Println("\n=== Part 3: AddTableOfContents ===")

	// Insert right after the Subtitle paragraph (index 2 in the body).
	target := doc.Paragraphs()[2]
	toc, err := doc.InsertTableOfContentsBefore(target, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Inserted TOC before %q (title: %q)\n",
		target.Text(), toc.Title().Text())

	// A second TOC variant at the end with a custom title and depth.
	_, err = doc.AddTableOfContents(&wordingo.TOCOptions{
		Title:  "Summary",
		Levels: 2,
	})
	if err != nil {
		log.Fatal(err)
	}

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Saved %s\n", out)

	// ============================================================
	// Part 4: Reopen and verify
	// ============================================================
	fmt.Println("\n=== Part 4: Reopen verification ===")

	re, err := wordingo.Open(out)
	if err != nil {
		log.Fatal(err)
	}
	defer re.Close()

	text, _ := re.ExtractText(nil)
	checks := map[string]string{
		"TOC title at start":   "Table of Contents",
		"Summary at end":       "Summary",
		"heading in TOC range": "Executive Summary",
	}
	for label, want := range checks {
		if contains(text, want) {
			fmt.Printf("  PASS: %s\n", label)
		} else {
			fmt.Printf("  FAIL: %s — %q not found\n", label, want)
		}
	}
	// The outline still reflects only real headings.
	headings := re.Headings()
	fmt.Printf("  Reopened outline has %d headings (TOC entries are not headings)\n", len(headings))
	if len(headings) != 5 {
		fmt.Println("  FAIL: expected 5 headings")
	} else {
		fmt.Println("  PASS: outline count correct")
	}
	if len(re.Warnings()) == 0 {
		fmt.Println("  PASS: no warnings")
	} else {
		fmt.Printf("  WARN: %v\n", re.Warnings())
	}

	fmt.Println("\n=== Done ===")
	fmt.Println("Open output.docx in Word: the TOC fields update automatically on open,")
	fmt.Println("so page numbers and entries refresh from the headings.")
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
