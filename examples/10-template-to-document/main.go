package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/fabiomarini/wordingo"
)

func testdataPath(parts ...string) string {
	_, filename, _, _ := runtime.Caller(0)
	base := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	return filepath.Join(append([]string{base, "testdata", "word"}, parts...)...)
}

func outputPath(name string) string {
	return filepath.Join(".", name)
}

func main() {
	// --- Example 1: FromTemplate — new doc with template styles ---
	fmt.Println("=== 1. FromTemplate: Fresh document with template styles ===")

	customStylesPath := testdataPath("03-custom_styles.docx")
	fmt.Printf("  Template styles: ")
	printStyleIDs(customStylesPath)

	doc1, err := wordingo.FromTemplate(customStylesPath)
	if err != nil {
		log.Fatalf("FromTemplate: %v", err)
	}
	defer doc1.Close()

	doc1.AddParagraph("Q3 Financial Report").SetStyle("Titolo")
	doc1.AddParagraph("Prepared by the Finance Team").SetStyle("Sottotitolo")

	doc1.AddParagraph("Executive Summary").SetStyle("Titolo1")
	doc1.AddParagraph("Revenue grew 23% year-over-year driven by strong performance in the Enterprise and SMB segments. Operating margins improved to 18.5%.")

	doc1.AddParagraph("Key Metrics").SetStyle("Titolo1")
	doc1.AddTable([][]string{
		{"Metric", "Q3 2025", "Q3 2026", "Change"},
		{"Revenue", "$4.2M", "$5.2M", "+23%"},
		{"Gross Margin", "72%", "74%", "+2pp"},
		{"Net Income", "$0.7M", "$0.96M", "+37%"},
		{"Active Customers", "1,240", "1,580", "+27%"},
	})

	doc1.AddParagraph("Regional Breakdown").SetStyle("Titolo2")
	doc1.AddParagraph("North America contributed 52% of revenue (up from 48%), EMEA 30%, and APAC 18%.")

	out1 := outputPath("from-template-report.docx")
	if err := doc1.Save(out1); err != nil {
		log.Fatal(err)
	}
	fi1, _ := os.Stat(out1)
	fmt.Printf("  Created %s (%d bytes)\n", out1, fi1.Size())

	for _, w := range doc1.Warnings() {
		fmt.Printf("  Warning: %s\n", w)
	}

	// --- Example 2: OpenTemplate — edit existing content with styles ---
	fmt.Println("\n=== 2. OpenTemplate: Edit existing styled document ===")

	samplePath := testdataPath("04-custom_styles_plus_sample_text.docx")
	doc2, err := wordingo.OpenTemplate(samplePath)
	if err != nil {
		log.Fatalf("OpenTemplate: %v", err)
	}
	defer doc2.Close()

	fmt.Printf("  Source has %d paragraphs\n", len(doc2.Paragraphs()))
	for i, p := range doc2.Paragraphs() {
		fmt.Printf("    [%d] style=%q text=%q\n", i, p.Style(), truncate(p.Text(), 60))
	}

	doc2.AddParagraph("Appendix: Updated Projections").SetStyle("Titolo1")

	doc2.AddParagraph("Following the quarterly review, forward guidance has been revised upward. Projected Q4 revenue is $5.5M with full-year 2026 revenue of $20M.")

	tbl := doc2.AddTableBuilder()
	tbl.SetTableStyle("LightGridAccent1")
	tbl.Row(0).Cell(0).SetText("Quarter").SetBold(true).SetShading("", "D9E2F3")
	tbl.Row(0).Cell(1).SetText("Projected Revenue").SetBold(true).SetShading("", "D9E2F3")
	tbl.Row(0).Cell(2).SetText("Confidence").SetBold(true).SetShading("", "D9E2F3")
	tbl.Row(0).Cell(3).SetText("Notes").SetBold(true).SetShading("", "D9E2F3")
	tbl.Row(1).Cell(0).SetText("Q4 2026")
	tbl.Row(1).Cell(1).SetText("$5.5M")
	tbl.Row(1).Cell(2).SetText("High")
	tbl.Row(1).Cell(3).SetText("Holiday season + new product launch")
	tbl.Row(2).Cell(0).SetText("Q1 2027")
	tbl.Row(2).Cell(1).SetText("$4.8M")
	tbl.Row(2).Cell(2).SetText("Medium")
	tbl.Row(2).Cell(3).SetText("Seasonal dip, offset by Enterprise renewals")

	out2 := outputPath("open-template-appendix.docx")
	if err := doc2.Save(out2); err != nil {
		log.Fatal(err)
	}
	fi2, _ := os.Stat(out2)
	fmt.Printf("  Created %s (%d bytes)\n", out2, fi2.Size())

	// --- Example 3: FromTemplate with Merge + Edit (Phase 6) ---
	fmt.Println("\n=== 3. FromTemplate + Merge + Edit combined ===")

	doc3, err := wordingo.FromTemplate(customStylesPath)
	if err != nil {
		log.Fatalf("FromTemplate: %v", err)
	}
	defer doc3.Close()

	h := doc3.AddHeader(wordingo.HeaderDefault)
	h.AddParagraph("Invoice {{invoice_id}}")

	doc3.AddParagraph("Invoice").SetStyle("Titolo")
	doc3.AddParagraph("Bill To: {{customer_name}}").SetStyle("Sottotitolo")

	doc3.AddParagraph("Items").SetStyle("Titolo1")
	doc3.AddTable([][]string{
		{"#", "Description", "Qty", "Unit Price", "Total"},
		{"1", "{{item_1}}", "{{qty_1}}", "{{price_1}}", "{{line_1}}"},
		{"2", "{{item_2}}", "{{qty_2}}", "{{price_2}}", "{{line_2}}"},
	})

	doc3.Merge(map[string]string{
		"invoice_id":    "INV-2026-0501",
		"customer_name": "Acme Corp",
		"item_1":        "Consulting services", "qty_1": "40", "price_1": "$150", "line_1": "$6,000",
		"item_2":        "Software license",    "qty_2": "5",  "price_2": "$500", "line_2": "$2,500",
	}, nil)

	var totalPara *wordingo.Paragraph
	for _, p := range doc3.Paragraphs() {
		if p.Text() == "" {
			totalPara = p
			break
		}
	}
	if totalPara != nil {
		doc3.InsertBefore(totalPara, "Total Due: $8,500")
	}

	out3 := outputPath("template-merge-invoice.docx")
	if err := doc3.Save(out3); err != nil {
		log.Fatal(err)
	}
	fi3, _ := os.Stat(out3)
	fmt.Printf("  Created %s (%d bytes)\n", out3, fi3.Size())

	for _, w := range doc3.Warnings() {
		fmt.Printf("  Warning: %s\n", w)
	}

	// --- Summary ---
	fmt.Println("\n=== Summary ===")
	fmt.Printf("  FromTemplate: %s\n", out1)
	fmt.Printf("  OpenTemplate: %s\n", out2)
	fmt.Printf("  Template+Merge: %s\n", out3)
}

func printStyleIDs(path string) {
	r, err := zip.OpenReader(path)
	if err != nil {
		fmt.Printf("  (cannot open: %v)\n", err)
		return
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name != "word/styles.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		re := regexp.MustCompile(`w:styleId="([^"]+)"`)
		matches := re.FindAllStringSubmatch(string(data), -1)
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m[1])
		}
		fmt.Printf("%v\n", ids)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
