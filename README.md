# WordInGo
![WordInGo logo](./docs/wordInGo.png)

Pure Go library for creating and editing Word documents — zero external dependencies.

## What it is

wordingo lets you build .docx files from scratch, open and edit existing ones, and convert between Word and Markdown. Everything is pure Go — no C bindings, no system calls to Word, no binaries to install. Just `go get` and you're off.

```go
doc, _ := wordingo.Create()
doc.AddParagraph("Hello, Word!")
doc.Save("output.docx")
```

## Key features

| What | How |
|------|-----|
| **Create & open** | Blank documents with default styles, or open existing .docx files and edit them |
| **Templates** | Clone a template's look (colors, fonts, heading styles) into your document |
| **Text & formatting** | Bold, italic, underline, color, highlights, font family, size — the works |
| **Tables** | Quick grid from `[][]string`, or a full builder with borders, shading, merged cells |
| **Images** | Embed PNG or JPEG photos; set display size in inches; auto DPI detection |
| **Headers & footers** | Default, first-page, and even-page variants with full paragraph editing |
| **Lists** | Bulleted, numbered, nested up to 9 levels |
| **Hyperlinks** | Clickable links with styled, colored text |
| **Page setup** | Portrait/landscape, Letter/A4/Legal, custom margins and page breaks |
| **Template merge** | Replace `{{placeholders}}` in text, table cells, headers, and footers |
| **Edit operations** | Insert or delete paragraphs by reference, replace text in runs, delete table rows |
| **Table of contents** | Generate a TOC from the document's own headings, with live field updates in Word |
| **Text extraction** | Pull plain text from the whole document or specific sections |
| **Markdown export** | Convert any .docx to GitHub-flavored Markdown |
| **Markdown import** | Create a Word doc from Markdown — headings, lists, tables, code blocks, images |
| **Inline docs** | No magic numbers: every unit type (`twips`, `dxa`, half-points) follows Word's conventions |

## Quick start

```go
package main

import (
    "log"
    "github.com/fabiomarini/wordingo"
)

func main() {
    doc, err := wordingo.Create()
    if err != nil { log.Fatal(err) }
    defer doc.Close()

    doc.AddParagraph("The Art of Go").SetStyle("Title")
    p := doc.AddParagraph("Go is statically typed, compiled, and fast.")
    p.AddRun(" Fast.").SetBold(true).SetColor("2E75B6")

    _, _ = doc.AddTable([][]string{
        {"Language", "Typing"},
        {"Go", "static"},
        {"Python", "dynamic"},
    })

    if err := doc.Save("output.docx"); err != nil { log.Fatal(err) }
}
```

## What you can do

### Create or open a document

```go
doc, _ := wordingo.Create()                 // blank doc — ready to write
defer doc.Close()
doc.Save("output.docx")                      // write to file
doc.WriteTo(w)                              // or any io.Writer

doc2, _ := wordingo.Open("existing.docx")   // open and edit
doc3, _ := wordingo.OpenReader(r, size)     // from io.ReaderAt
```

### Use a template

```go
// Styles from template, write your own content:
doc, _ := wordingo.FromTemplate("corporate-template.docx")

// Keep the template's content and add more:
doc, _ := wordingo.OpenTemplate("letterhead.docx")
```

`FromTemplate` copies all the formatting (styles, fonts, colors, numbering) into an empty document — great when you want the look but write everything yourself. `OpenTemplate` keeps the original paragraphs so you can append or edit.

### Write paragraphs and style text

```go
p := doc.AddParagraph("Chapter 1").SetStyle("Heading1")
p.AddRun(" — A thrilling start.")
r := p.AddRun("Bold and blue").SetBold(true).SetColor("2E75B6")
r.ReplaceText("blue", "red")
```

Paragraph methods: `SetAlignment`, `SetSpacing`, `SetIndent`, `SetPageBreakBefore`. Run methods: `SetBold`, `SetItalic`, `SetUnderline`, `SetFont`, `SetSize`, `SetColor`, `SetHighlight`, `SetStyle`.

### Style names

Blank documents include a full set of default styles. Available style IDs include `"Title"`, `"Subtitle"`, `"Heading1"`–`"Heading9"`, `"Normal"`, `"Quote"`, `"IntenseEmphasis"`, `"IntenseReference"`, `"ListParagraph"`, and character-style variants.

```go
doc.AddParagraph("Welcome").SetStyle("Title")
p.AddRun("important").SetStyle("IntenseEmphasis")
```

### Create tables

```go
// Quick table from a spreadsheet-style grid:
doc.AddTable([][]string{{"Item", "Price"}, {"Widget", "$5"}})

// Full control:
tbl := doc.AddTableBuilder()
tbl.SetTableStyle("LightGridAccent1")
tbl.SetWidth(8000, "dxa")
tbl.Row(0).Cell(0).SetText("Product").SetBold(true)
tbl.Row(0).Cell(0).MergeRight()             // span 2 columns
if err := tbl.DeleteRow(2); err != nil { /* OOB */ }
```

Tables are added after all paragraphs (current behavior). Each cell supports text, bold, shading, width, and merging.

### Insert images

```go
run, _ := doc.AddImage("photo.png")          // loads from disk
run, _ := doc.AddImageBytes("gradient.png", data, "image/png")
run.SetImageWidth(3.0).SetImageHeight(2.0)    // display size in inches
```

PNG and JPEG both work. DPI is detected automatically when present in the file; otherwise defaults to 72 DPI with a 3-inch width.

### Add headers and footers

```go
h := doc.AddHeader(wordingo.HeaderDefault)
h.AddParagraph("Confidential")
f := doc.AddFooter(wordingo.FooterDefault)
f.AddParagraph("Page ")
```

Headers come in three variants: default, first-page only, and even-page. Footers have the same options. All support editing paragraphs after creation.

### Bulleted and numbered lists

```go
doc.AddListFromSlice([]string{"Red", "Green", "Blue"}, false)   // bulleted
doc.AddListFromSlice([]string{"First", "Second"}, true)         // numbered

lb := doc.AddList(true)
lb.AddItem("Item 1", 0)
lb.AddItem("Sub-item", 1).AddItem("Sub-sub", 2)                 // 9 nesting levels

doc.AddNumberingDef("upperRoman", 1)                            // custom numbering
```

Lists handle nesting up to 9 levels deep. Custom numbering formats (roman numerals, letters, etc.) are supported.

### Hyperlinks

```go
p := doc.AddParagraph("Visit ")
r := p.AddHyperlink("the docs", "https://go.dev")
r.SetColor("0563C1").SetUnderline("single")
```

Each hyperlink creates a clickable link in the Word document. The returned run accepts all the usual formatting methods.

### Page layout

```go
doc.SetOrientation(wordingo.OrientationLandscape)
doc.SetPaperSize(wordingo.PaperA4W, wordingo.PaperA4H)
doc.SetMargins(1440, 1440, 1440, 1440)          // top, right, bottom, left (in twips)
doc.AddPageBreak()

doc.Section().SetOrientation(wordingo.OrientationLandscape)     // or via Section wrapper
```

Available paper sizes: `PaperLetter`, `PaperA4`, `PaperLegal` — or set custom dimensions.

### Replace placeholders (template merge)

```go
doc.Merge(map[string]string{
    "name":  "Alice",
    "item":  "invoice",
}, nil)                                             // all sections

doc.Merge(map[string]string{"page":"2"}, &wordingo.MergeOpts{
    ScopedParts: wordingo.ScopedParts{Headers: true, Footers: true},
})
```

Replaces `{{name}}` and `{{item}}` wherever they appear — body text, table cells, headers, footers. Works even if placeholders get split across multiple formatting runs. Unused keys surface as warnings.

### Insert, delete, and edit content

```go
target := doc.Paragraphs()[2]
doc.InsertBefore(target, "Start here")               // insert above a paragraph
doc.InsertAfter(target,  "End here")                 // insert below
doc.DeleteParagraph(target)                          // remove a paragraph
r.SetText("replacement")
r.ReplaceText("old", "new")
tbl.DeleteRow(1)                                     // remove a table row

// Iterate mixed paragraphs and tables in document order:
for _, el := range doc.Body() {
    switch el.Type {
    case wordingo.ElementParagraph: /* el.Para */
    case wordingo.ElementTable:     /* el.Table */
    }
}
```

### Extract text

```go
text, _ := doc.ExtractText(nil)                       // all content, newline-separated
text, _ := doc.ExtractText(&wordingo.ExtractOpts{
    ScopedParts: wordingo.ScopedParts{Body: true},
    Separator:   " | ",
})
```

Extract text from the body, tables, headers, and footers — or pick specific sections. Read-only, doesn't modify the document.

### Table of contents / summary

Build a table of contents straight from the document's own headings:

```go
// Inspect the outline first, if you like:
for _, h := range doc.Headings() {
    fmt.Printf("%d. %s\n", h.Level, h.Text)   // 1. Introduction
}

// Append a TOC section at the end of the document:
toc, _ := doc.AddTableOfContents(nil)
toc.Title().SetStyle("Title")                 // restyle the heading if you want

// Or place it before a specific paragraph (e.g. right after the title):
doc.InsertTableOfContentsBefore(doc.Paragraphs()[1], &wordingo.TOCOptions{
    Title:  "Contents",
    Levels: 2,
})
```

The generated section is a real Word TOC field (`TOC \o "1-N" \h \z \u`): it collects every body paragraph styled `Heading1`–`Heading9` (or carrying an explicit outline level). Entries link to bookmarks placed on the headings, and `w:updateFields` is set in the document settings so Word repopulates the TOC — with live page numbers — the moment the file is opened. Viewers that don't refresh fields still see the generated summary. Set `TOCOptions.UpdateOnOpen` to `false` to leave the document settings untouched.

### Convert to and from Markdown

```go
md, _ := doc.ToMarkdown(nil)                        // .docx → GitHub-flavored Markdown

doc, _ := wordingo.CreateFromMarkdown("# Hi\n\n- a\n- b\n")   // Markdown → .docx
doc.ImportMarkdown("## Append more\n\nExtra content")         // append to existing doc
```

Export handles headings, bold, italic, tables, links, images, lists, and code blocks. Import handles all the same — you can round-trip content between Word and Markdown.

### Warnings & low-level access

```go
for _, w := range doc.Warnings() { fmt.Println(w) }   // non-critical issues

// Get the underlying XML struct when you need something the public API
// doesn't cover:
pkg := doc.X()              // document internals
ctP  := para.X()            // raw paragraph XML
ctR  := run.X()             // raw run XML
```

`Warnings()` flags issues like unknown style names or invalid colors — non-fatal, but useful for catching mistakes. The `X()` methods give you direct access to the library's internal XML structs, handy for edge cases the high-level API doesn't address.

## Examples

Each example is a standalone Go program. Run it from its directory:

```bash
cd examples/01-blank-doc && go run main.go
```

| # | Example | What it shows |
|---|---------|---------------|
| 1 | [01-blank-doc](examples/01-blank-doc/README.md) | Minimum "Hello World" document |
| 2 | [02-text-and-styles](examples/02-text-and-styles/README.md) | Paragraphs, headings, inline formatting |
| 3 | [03-tables](examples/03-tables/README.md) | Tables — quick grid and builder |
| 4 | [04-images](examples/04-images/README.md) | Embedded PNG images |
| 5 | [05-headers-footers](examples/05-headers-footers/README.md) | Headers, footers, A4, landscape |
| 6 | [06-lists](examples/06-lists/README.md) | Bulleted, ordered, nested lists |
| 7 | [07-hyperlinks](examples/07-hyperlinks/README.md) | Clickable links |
| 8 | [08-comprehensive](examples/08-comprehensive/README.md) | All features in one document |
| 9 | [09-merge-and-edit](examples/09-merge-and-edit/README.md) | Placeholder merge, insert/delete, edit |
| 10 | [10-template-to-document](examples/10-template-to-document/README.md) | Using templates |
| 11 | [11-text-extraction-markdown](examples/11-text-extraction-markdown/README.md) | Text extraction, markdown round-trip |
| 12 | [12-table-of-contents](examples/12-table-of-contents/README.md) | Table of contents from the document's headings |

## Design principles

- **Zero dependencies.** Only the Go standard library. `go.mod` has no `require` lines.
- **Faithful to the spec.** Built on the ISO OOXML standard. Everything you don't explicitly touch stays as-is — opening and saving a document changes nothing you didn't ask for.
- **Fidelity first.** Templates are copied byte-for-byte, never re-encoded. Your template's styles, fonts, and colors come through exactly as designed.
- **Opens clean in Word.** The test is simple: does the output open in Microsoft Word without a repair prompt?
- **Escape hatch when you need it.** If the public API doesn't expose some XML attribute, `X()` gives you direct access. No need to wait for a wrapper.

## Installation

```bash
go get github.com/fabiomarini/wordingo
```

Requires Go 1.23+.

## Status

v0.1.0 covers: create/open/save, templates, paragraphs and runs, named styles, formatting, tables, images, headers/footers, lists, hyperlinks, page setup, template merge, edit operations, text extraction, and Markdown round-trip.

## License

MIT, copyright 2026 Fabio Marini. See [LICENSE](LICENSE).

## Contributing

PRs welcome. For larger changes, open an issue first. No external dependencies allowed — stdlib only. Run `go test ./...` before submitting.
