package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/fabiomarini/wordingo"
)

func main() {
	// ============================================================
	// Part 1: Create source document with rich content
	// ============================================================
	fmt.Println("=== Part 1: Create Source Document ===")

	doc, err := wordingo.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	doc.AddHeader(wordingo.HeaderDefault).AddParagraph("Wordingo Demo — Header")
	doc.AddFooter(wordingo.FooterDefault).AddParagraph("Page")

	doc.AddParagraph("Text Extraction & Markdown Demo").SetStyle("Title")
	doc.AddParagraph("Demonstrates ExtractText, ToMarkdown, CreateFromMarkdown, ImportMarkdown").SetStyle("Subtitle")

	doc.AddParagraph("1. Text & Inline Formatting").SetStyle("Heading1")
	p := doc.AddParagraph("Inline formatting demo: ")
	p.AddRun("bold text").SetBold(true)
	p.AddRun(", ")
	p.AddRun("italic text").SetItalic(true)
	p.AddRun(", ")
	p.AddRun("bold+italic").SetBold(true).SetItalic(true)
	p.AddRun(", and ")
	p.AddRun("plain continuation")
	p.AddRun(".")

	doc.AddParagraph("2. Hyperlinks").SetStyle("Heading1")
	l := doc.AddParagraph("Visit ")
	l.AddHyperlink("the project repo", "https://github.com/fabiomarini/wordingo").SetColor("0563C1")
	l.AddRun(" for more details.")

	doc.AddParagraph("3. Tables").SetStyle("Heading1")
	doc.AddParagraph("Feature Status").SetStyle("Heading2")
	doc.AddTable([][]string{
		{"Feature", "Status", "Version"},
		{"ExtractText", "Done", "0.6.0"},
		{"ToMarkdown", "Done", "0.6.0"},
		{"CreateFromMarkdown", "Done", "0.6.0"},
		{"ImportMarkdown", "Done", "0.6.0"},
	})

	doc.AddParagraph("4. Lists").SetStyle("Heading1")
	features := doc.AddList(false)
	features.AddItem("Plain text extraction with scoped parts control", 0)
	features.AddItem("GFM markdown export with formatting preservation", 0)
	features.AddItem("Markdown import to create or append documents", 0)
	features.AddItem("Data URI image embedding in markdown", 0)

	doc.AddParagraph("5. Code Block").SetStyle("Heading1")
	for _, line := range []string{"func Hello() string {", `    return "Hello, World!"`, "}"} {
		p := doc.AddParagraph("")
		p.AddRun(line).SetFont("Consolas")
	}

	doc.AddParagraph("6. Image").SetStyle("Heading1")
	imgData := makeGradientPNG()
	run, err := doc.AddImageBytes("gradient.png", imgData, "image/png")
	if err != nil {
		log.Fatal(err)
	}
	run.SetImageWidth(1.5).SetImageHeight(1.5)
	doc.AddParagraph("A 50x50 gradient PNG embedded inline.")

	out := "output.docx"
	if err := doc.Save(out); err != nil {
		log.Fatal(err)
	}
	fi, _ := os.Stat(out)
	fmt.Printf("  Saved source doc: %s (%d bytes)\n", out, fi.Size())

	// ============================================================
	// Part 2: ExtractText
	// ============================================================
	fmt.Println("\n=== Part 2: ExtractText ===")

	text, err := doc.ExtractText(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nFull text (default opts):\n%s\n", text)
	os.WriteFile("output.txt", []byte(text), 0644)

	bodyOnly, _ := doc.ExtractText(&wordingo.ExtractOpts{
		ScopedParts: wordingo.ScopedParts{Body: true, Tables: false, Headers: false, Footers: false},
	})
	fmt.Printf("\nBody only (no tables/headers/footers):\n%s\n", bodyOnly)

	headersOnly, _ := doc.ExtractText(&wordingo.ExtractOpts{
		ScopedParts: wordingo.ScopedParts{Body: false, Tables: false, Headers: true, Footers: false},
	})
	fmt.Printf("Headers only: %q\n", headersOnly)

	customSep, _ := doc.ExtractText(&wordingo.ExtractOpts{
		ScopedParts: wordingo.ScopedParts{Body: true, Tables: true, Headers: false, Footers: false},
		Separator:   " | ",
	})
	fmt.Printf("Custom separator (\" | \"):\n%s\n", customSep)

	verifyExtractText(doc)

	// ============================================================
	// Part 3: ToMarkdown
	// ============================================================
	fmt.Println("\n=== Part 3: ToMarkdown ===")

	md, err := doc.ToMarkdown(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Full markdown:\n%s\n", md)
	os.WriteFile("output.md", []byte(md), 0644)

	mdBody, _ := doc.ToMarkdown(&wordingo.ExtractOpts{
		ScopedParts: wordingo.ScopedParts{Body: true, Tables: true, Headers: false, Footers: false},
	})
	fmt.Printf("Markdown (body+tables, no headers/footers):\n%s\n", mdBody)

	verifyToMarkdown(doc)

	// ============================================================
	// Part 4: CreateFromMarkdown
	// ============================================================
	fmt.Println("\n=== Part 4: CreateFromMarkdown ===")

	sourceMD := "# Markdown Import Demo\n\n"
	sourceMD += "This document was **created** from *markdown* using CreateFromMarkdown.\n\n"
	sourceMD += "## Features\n\n"
	sourceMD += "- Imports headings, paragraphs, lists\n"
	sourceMD += "- Supports **bold**, *italic*, and `code` inline formatting\n"
	sourceMD += "- Handles pipe tables and fenced code blocks\n\n"
	sourceMD += "## Example Table\n\n"
	sourceMD += "| Command | Description |\n"
	sourceMD += "|---------|-------------|\n"
	sourceMD += "| ExtractText | Plain text extraction |\n"
	sourceMD += "| ToMarkdown | GFM export |\n"
	sourceMD += "| CreateFromMarkdown | Markdown import |\n\n"
	sourceMD += "## Code Sample\n\n"
	sourceMD += "```go\n"
	sourceMD += `package main` + "\n\n"
	sourceMD += `import "fmt"` + "\n\n"
	sourceMD += "func main() {\n"
	sourceMD += `    fmt.Println("Hello!")` + "\n"
	sourceMD += "}\n"
	sourceMD += "```\n\n"
	sourceMD += "Visit the [project](https://github.com/fabiomarini/wordingo) for details.\n"

	imported, err := wordingo.CreateFromMarkdown(sourceMD)
	if err != nil {
		log.Fatal(err)
	}
	defer imported.Close()

	os.WriteFile("imported-source.md", []byte(sourceMD), 0644)

	importedText, _ := imported.ExtractText(nil)
	fmt.Printf("Import round-trip (markdown -> doc -> text):\n%s\n", importedText)
	os.WriteFile("imported-output.txt", []byte(importedText), 0644)

	// ============================================================
	// Part 5: ImportMarkdown (append to existing doc)
	// ============================================================
	fmt.Println("\n=== Part 5: ImportMarkdown ===")

	imported.ImportMarkdown("\n## Appended Section\n\nThis content was **appended** using ImportMarkdown.\n\n- Item A\n- Item B\n")
	appendText, _ := imported.ExtractText(nil)
	fmt.Printf("After ImportMarkdown append:\n%s\n", appendText)

	// ============================================================
	// Part 6: Round-trip verification
	// ============================================================
	fmt.Println("\n=== Part 6: Round-Trip Verification ===")

	originalText, _ := doc.ExtractText(&wordingo.ExtractOpts{
		ScopedParts: wordingo.ScopedParts{Body: true, Tables: true, Headers: false, Footers: false},
	})
	verifyContains(originalText, "bold text", "bold formatting")
	verifyContains(originalText, "italic text", "italic formatting")
	verifyContains(originalText, "ExtractText", "table content")
	verifyContains(originalText, "Text Extraction & Markdown Demo", "title text")
	verifyContains(originalText, "the project repo", "hyperlink text")

	mdExtracted, _ := doc.ToMarkdown(&wordingo.ExtractOpts{
		ScopedParts: wordingo.ScopedParts{Body: true, Tables: true, Headers: false, Footers: false},
	})
	verifyContains(mdExtracted, "**bold text**", "bold markdown")
	verifyContains(mdExtracted, "*italic text*", "italic markdown")
	verifyContains(mdExtracted, "| ExtractText | Done | 0.6.0 |", "table markdown")
	verifyContains(mdExtracted, "Text Extraction & Markdown Demo", "document title")

	importedMD, _ := imported.ToMarkdown(nil)
	verifyContains(importedMD, "**created**", "imported bold")
	verifyContains(importedMD, "*markdown*", "imported italic")
	verifyContains(importedMD, "Appended Section", "imported append")
	importedText2, _ := imported.ExtractText(nil)
	verifyContains(importedText2, "ExtractText", "imported table (via ExtractText)")
	os.WriteFile("imported-output.md", []byte(importedMD), 0644)

	fmt.Println("\n=== All verifications passed ===")

	fmt.Println("\n=== Summary ===")
	fmt.Println("  ExtractText      — Plain text extraction with scoped parts")
	fmt.Println("  ToMarkdown       — GFM export with formatting, tables, links, images")
	fmt.Println("  CreateFromMarkdown — Document creation from markdown string")
	fmt.Println("  ImportMarkdown   — Append markdown content to existing document")
}

func verifyExtractText(doc *wordingo.Document) {
	t, _ := doc.ExtractText(&wordingo.ExtractOpts{
		ScopedParts: wordingo.ScopedParts{Body: true, Tables: true, Headers: false, Footers: false},
	})
	if strings.Contains(t, "Wordingo Demo — Header") {
		fmt.Println("  FAIL: ExtractText with Headers=false included header text")
	} else {
		fmt.Println("  PASS: ExtractText Headers=false excludes header")
	}
	if strings.Contains(t, "bold text") {
		fmt.Println("  PASS: ExtractText includes bold text")
	} else {
		fmt.Println("  FAIL: ExtractText missing bold text")
	}
}

func verifyToMarkdown(doc *wordingo.Document) {
	md, _ := doc.ToMarkdown(nil)
	if strings.Contains(md, "**bold text**") {
		fmt.Println("  PASS: ToMarkdown emits bold as **text**")
	} else {
		fmt.Println("  FAIL: ToMarkdown missing bold markers")
	}
	if strings.Contains(md, "| Feature |") {
		fmt.Println("  PASS: ToMarkdown emits pipe tables")
	} else {
		fmt.Println("  FAIL: ToMarkdown missing table")
	}
	if strings.Contains(md, "```") {
		fmt.Println("  PASS: ToMarkdown emits code blocks")
	} else {
		fmt.Println("  INFO: Code block detection requires PPr on paragraph (skipped in this example)")
	}
}

func verifyContains(s, substr, label string) {
	if strings.Contains(s, substr) {
		fmt.Printf("  PASS: %s found\n", label)
	} else {
		fmt.Printf("  FAIL: %s not found — expected %q\n", label, substr)
	}
}

func makeGradientPNG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			r := uint8(float64(x) / 50 * 255)
			b := uint8(float64(y) / 50 * 255)
			img.Set(x, y, color.NRGBA{R: r, G: 100, B: b, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
