package markdown

import (
	"strings"
	"testing"
)

func TestParseEmpty(t *testing.T) {
	blocks, err := Parse("")
	if err != nil {
		t.Fatalf("Parse empty: unexpected error: %v", err)
	}
	if len(blocks) != 0 {
		t.Fatalf("Parse empty: expected 0 blocks, got %d", len(blocks))
	}
}

func TestParseHeading(t *testing.T) {
	for lvl := 1; lvl <= 6; lvl++ {
		input := strings.Repeat("#", lvl) + " Heading"
		blocks, err := Parse(input)
		if err != nil {
			t.Fatalf("Parse heading level %d: %v", lvl, err)
		}
		if len(blocks) != 1 {
			t.Fatalf("Parse heading level %d: expected 1 block, got %d", lvl, len(blocks))
		}
		if blocks[0].Type != BlockHeading {
			t.Fatalf("Parse heading level %d: expected BlockHeading", lvl)
		}
		if blocks[0].Level != lvl {
			t.Fatalf("Parse heading level %d: expected level %d, got %d", lvl, lvl, blocks[0].Level)
		}
		if blocks[0].Content != "Heading" {
			t.Fatalf("Parse heading level %d: expected 'Heading', got %q", lvl, blocks[0].Content)
		}
	}
}

func TestParseParagraph(t *testing.T) {
	input := "Hello world"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse paragraph: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse paragraph: expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockParagraph {
		t.Fatalf("Parse paragraph: expected BlockParagraph")
	}
	if blocks[0].Content != "Hello world" {
		t.Fatalf("Parse paragraph: expected 'Hello world', got %q", blocks[0].Content)
	}
}

func TestParseCodeBlock(t *testing.T) {
	input := "```go\npackage main\n```"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse code block: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse code block: expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockCodeBlock {
		t.Fatalf("Parse code block: expected BlockCodeBlock")
	}
	if blocks[0].Language != "go" {
		t.Fatalf("Parse code block: expected Language 'go', got %q", blocks[0].Language)
	}
	if len(blocks[0].Lines) != 1 || blocks[0].Lines[0] != "package main" {
		t.Fatalf("Parse code block: expected lines ['package main'], got %v", blocks[0].Lines)
	}
}

func TestParseCodeBlockNoLang(t *testing.T) {
	input := "```\ncode\n```"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse code block no lang: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse code block no lang: expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockCodeBlock {
		t.Fatalf("Parse code block no lang: expected BlockCodeBlock")
	}
	if blocks[0].Language != "" {
		t.Fatalf("Parse code block no lang: expected empty language, got %q", blocks[0].Language)
	}
}

func TestParseBulletList(t *testing.T) {
	input := "- item1\n- item2"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse bullet list: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse bullet list: expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockBulletList {
		t.Fatalf("Parse bullet list: expected BlockBulletList")
	}
	if len(blocks[0].ListItems) != 2 {
		t.Fatalf("Parse bullet list: expected 2 items, got %d", len(blocks[0].ListItems))
	}
	if blocks[0].ListItems[0].Content != "item1" {
		t.Fatalf("Parse bullet list: first item expected 'item1', got %q", blocks[0].ListItems[0].Content)
	}
	if blocks[0].ListItems[1].Content != "item2" {
		t.Fatalf("Parse bullet list: second item expected 'item2', got %q", blocks[0].ListItems[1].Content)
	}
}

func TestParseOrderedList(t *testing.T) {
	input := "1. first\n2. second"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse ordered list: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse ordered list: expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockOrderedList {
		t.Fatalf("Parse ordered list: expected BlockOrderedList")
	}
	if len(blocks[0].ListItems) != 2 {
		t.Fatalf("Parse ordered list: expected 2 items, got %d", len(blocks[0].ListItems))
	}
}

func TestParsePipeTable(t *testing.T) {
	input := "|a|b|\n|---|---|\n|1|2|"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse pipe table: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse pipe table: expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockTable {
		t.Fatalf("Parse pipe table: expected BlockTable")
	}
	if len(blocks[0].Cells) != 2 {
		t.Fatalf("Parse pipe table: expected 2 rows, got %d", len(blocks[0].Cells))
	}
	if len(blocks[0].Cells[0]) != 2 || blocks[0].Cells[0][0] != "a" || blocks[0].Cells[0][1] != "b" {
		t.Fatalf("Parse pipe table: header row mismatch, got %v", blocks[0].Cells[0])
	}
	if len(blocks[0].Cells[1]) != 2 || blocks[0].Cells[1][0] != "1" || blocks[0].Cells[1][1] != "2" {
		t.Fatalf("Parse pipe table: data row mismatch, got %v", blocks[0].Cells[1])
	}
}

func TestParseBoldInline(t *testing.T) {
	input := "**bold**"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse bold: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse bold: expected 1 block, got %d", len(blocks))
	}
	if len(blocks[0].Inlines) != 1 {
		t.Fatalf("Parse bold: expected 1 inline span, got %d", len(blocks[0].Inlines))
	}
	if !blocks[0].Inlines[0].Bold {
		t.Fatalf("Parse bold: expected Bold=true")
	}
	if blocks[0].Inlines[0].Text != "bold" {
		t.Fatalf("Parse bold: expected Text 'bold', got %q", blocks[0].Inlines[0].Text)
	}
}

func TestParseItalicInline(t *testing.T) {
	input := "*italic*"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse italic: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse italic: expected 1 block")
	}
	if len(blocks[0].Inlines) != 1 {
		t.Fatalf("Parse italic: expected 1 inline span")
	}
	if !blocks[0].Inlines[0].Italic {
		t.Fatalf("Parse italic: expected Italic=true")
	}
	if blocks[0].Inlines[0].Text != "italic" {
		t.Fatalf("Parse italic: expected Text 'italic'")
	}
}

func TestParseCodeSpanInline(t *testing.T) {
	input := "`code`"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse code span: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse code span: expected 1 block")
	}
	if len(blocks[0].Inlines) != 1 {
		t.Fatalf("Parse code span: expected 1 inline span")
	}
	if !blocks[0].Inlines[0].Code {
		t.Fatalf("Parse code span: expected Code=true")
	}
	if blocks[0].Inlines[0].Text != "code" {
		t.Fatalf("Parse code span: expected Text 'code'")
	}
}

func TestParseLinkInline(t *testing.T) {
	input := "[text](url)"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse link: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse link: expected 1 block")
	}
	if len(blocks[0].Inlines) != 1 {
		t.Fatalf("Parse link: expected 1 inline span")
	}
	if blocks[0].Inlines[0].LinkText != "text" {
		t.Fatalf("Parse link: expected LinkText 'text', got %q", blocks[0].Inlines[0].LinkText)
	}
	if blocks[0].Inlines[0].LinkURL != "url" {
		t.Fatalf("Parse link: expected LinkURL 'url', got %q", blocks[0].Inlines[0].LinkURL)
	}
}

func TestParseImageInline(t *testing.T) {
	input := "![alt](img.png)"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse image: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse image: expected 1 block")
	}
	if len(blocks[0].Inlines) != 1 {
		t.Fatalf("Parse image: expected 1 inline span")
	}
	if blocks[0].Inlines[0].ImageAlt != "alt" {
		t.Fatalf("Parse image: expected ImageAlt 'alt', got %q", blocks[0].Inlines[0].ImageAlt)
	}
	if blocks[0].Inlines[0].ImageURL != "img.png" {
		t.Fatalf("Parse image: expected ImageURL 'img.png', got %q", blocks[0].Inlines[0].ImageURL)
	}
}

func TestParseCombinedBoldItalic(t *testing.T) {
	input := "**bold** and *italic*"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse combined: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse combined: expected 1 block")
	}
	if len(blocks[0].Inlines) != 3 {
		t.Fatalf("Parse combined: expected 3 inline spans, got %d", len(blocks[0].Inlines))
	}
}

func TestParseOnlyWhitespace(t *testing.T) {
	input := "   "
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse whitespace: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("Parse whitespace: expected 1 block")
	}
	if blocks[0].Type != BlockParagraph {
		t.Fatalf("Parse whitespace: expected BlockParagraph")
	}
}

func TestParseCommentInline(t *testing.T) {
	input := "<!-- html -->"
	blocks, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse comment: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Type != BlockParagraph {
		t.Fatalf("Parse comment: expected BlockParagraph")
	}
}
