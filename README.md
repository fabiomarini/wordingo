# wordingo

Pure Go, zero-dependency library for creating and editing Microsoft Word (.docx) documents — built on the OOXML (WordprocessingML) standard.

## What it is

wordingo is a pure Go library (not a CLI) for constructing and round-tripping WordprocessingML (.docx) packages. It is MIT-licensed and uses only the Go standard library plus its own internal subpackages (`internal/opc`, `internal/wml`, `internal/xmlutil`, `internal/style`, `internal/markdown`); `go.mod` has no `require` directives. The module path is `github.com/fabiomarini/wordingo` and it targets Go 1.23.

You can create styled .docx documents from scratch, open existing .docx files for round-trip editing with zero unintended changes, or clone a template's style dependency graph (styles, numbering, font table, theme, settings) into either a fresh empty body or a preserved body. The validation bar throughout the project is: the output file opens in Microsoft Word without any repair prompt.

Around sixty essential WordprocessingML (WML) types are modeled as Go structs. Unknown XML elements are preserved verbatim via raw-element hoisting rather than lossy-parsed, so round-trips of producer-documents drop no data.

## Key features

- Create blank documents with default styles, theme, font table, settings (Letter, 1-inch margins)
- Open existing .docx and round-trip them with zero unintended changes
- Clone a template's entire style dependency graph (styles.xml, numbering.xml, fontTable.xml, theme.xml, settings.xml) into a fresh body via `FromTemplate`, or preserve body content via `OpenTemplate`
- Paragraphs, runs, and named styles (paragraph- and run-level)
- Rich inline formatting: bold, italic, underline, font, size (half-points), color (hex), highlight, style
- Tables: simple grid (`AddTable` from `[][]string`) and full fluent builder with style, width, borders, shading, per-cell shading/bold/width, `MergeRight`/`MergeDown`, `DeleteRow`
- Images: PNG + JPEG embedded as DrawingML inline; DPI detection from JFIF/EXIF/PNG pHYs; auto-size or explicit `SetImageWidth`/`SetImageHeight` in inches; locked aspect ratio, stretch-to-fill
- Headers and footers: default/first/even variants; OPC parts created lazily; header/footer paragraphs editable via `InsertParagraphAt`/`DeleteParagraphAt`/`AddParagraph`
- Ordered and bulleted lists with full 9-level nesting depth (levels 0-8); `AddList`, `AddListFromSlice`, `AddNumberingDef` (custom numFmt/start); numbering.xml merged with any template numbering
- Hyperlinks with fresh OPC relationships and `Run`-style chaining (`SetBold`/`SetColor`/`SetUnderline` on the returned `*Run`)
- Page setup: orientation (portrait/landscape), paper size (Letter/A4/Legal constants; custom twips), margins (twips), page breaks (`AddPageBreak` or `SetPageBreakBefore`); chainable on `Document`
- Template merge: `{{placeholder}}` replacement in body, table cells, headers, footers; handles split-run placeholders; scoped via `MergeOpts.ScopedParts`; warns on unused keys
- Edit operations: `InsertBefore`/`InsertAfter` by paragraph pointer identity, `DeleteParagraph` by identity, `Run.SetText`, `Run.ReplaceText`, `TableBuilder.DeleteRow`
- `Body()` ordered iteration mixing paragraphs and tables (`BodyElement.type` = `ElementParagraph` | `ElementTable`)
- Text extraction: `doc.ExtractText(ExtractOpts)` — plain text from body/tables/headers/footers, scoped, custom separator
- Markdown export: `doc.ToMarkdown(ExtractOpts)` — GFM output (headings, bold/italic, pipe tables, links, image data URIs, lists, code blocks)
- Markdown import: `CreateFromMarkdown(string)` builds a new `Document`; `doc.ImportMarkdown(string)` appends to an existing doc; supports headings, paragraphs, ordered/bulleted lists, pipe tables, fenced code blocks, inline bold/italic/code/links/data-URI images
- Validation surface: `doc.Warnings()` returns non-fatal issues (unknown style refs, invalid colors, negative sizes, unused merge keys, markdown parse errors, package-layer warnings)
- Escape hatch via `X()` methods on `Document`, `Paragraph`, `Run`, `TableBuilder`, `RowBuilder`, `CellBuilder`, `ListBuilder`, `Header`, `Footer`, `Section` for raw WML access

## Installation

```
go get github.com/fabiomarini/wordingo
```

Requires Go 1.23. Module path: `github.com/fabiomarini/wordingo`.

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

    doc.AddParagraph("The Art of Go").SetStyle("Titolo")
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

## Comprehensive feature walkthrough

### Creating & opening documents

```go
doc, _ := wordingo.Create()                 // blank doc, default styles, Letter, 1in margins
defer doc.Close()
_ = doc.Save("output.docx")                   // serialize to ZIP .docx
n, _ := doc.WriteTo(w)                       // write to any io.Writer; returns bytes written
_ = doc.SaveFile("output.docx")               // alternate file writer

doc2, _ := wordingo.Open("existing.docx")     // parse body eagerly, supporting parts lazy
defer doc2.Close()
doc3, _ := wordingo.OpenReader(r, size)       // io.ReaderAt variant of Open
```

### Templates

```go
// Fresh body, template's full style dependency graph cloned.
d1, _ := wordingo.FromTemplate("template.docx")
// Preserved body content (paragraphs + tables) plus cloned styles,
// with header/footer refs stripped to avoid dangling rIds.
d2, _ := wordingo.OpenTemplate("template.docx")
// io.ReaderAt variants:
d3, _ := wordingo.FromTemplateReader(r, size)
d4, _ := wordingo.OpenTemplateReader(r, size)
```

`FromTemplate` clones styles/numbering/fontTable/theme/settings into a fresh empty body — use it when you want template styles but write the body yourself. `OpenTemplate` keeps the template's existing paragraphs and tables so you can append/edit on top. Both clone the style dependency graph via byte pass-through (no reserialization of style parts).

### Paragraphs & runs

```go
p := doc.AddParagraph("Hello")                // appends a paragraph with one run
p.AddRun(" world")                            // append another run
p.Text()                                       // concatenated run text
p.Style()                                      // current paragraph style ID, "" if none
r := p.AddRun("hi")
r.SetText("hi there")                          // overwrite run text
r.ReplaceText("hi", "hello")                   // strings.ReplaceAll on run text
```

### Paragraph formatting

```go
p.SetAlignment(wordingo.AlignmentLeft)         // AlignmentLeft/Center/Right/Both
p.SetSpacing(&wordingo.ParSpacing{Before:240, After:240, Line:360, LineRule:"auto"})
p.SetIndent(&wordingo.ParIndent{Left:720, Right:360, FirstLine:360, Hanging:360})
p.SetPageBreakBefore(true)                     // paragraph starts on a new page
p.SetFormatting(wordingo.ParFormat{Alignment:&a, Spacing:s, Indent:i}) // bulk
```

### Run formatting

```go
r.SetBold(true)
r.SetItalic(true)
r.SetUnderline("single")                        // any WML underline val
r.SetFont("Consolas")
r.SetSize(12)                                   // points; stored as half-points
r.SetColor("2E75B6")                            // 6 hex digits, else warns
r.SetHighlight("yellow")
r.SetStyle("EmphasisIntense")                   // run-style reference
r.SetFormatting(wordingo.RunFormat{Bold:&yes, Color:&c}) // bulk
```

### Named styles

Blank documents created via `Create()` ship with **Italian-named** style IDs from `defaults/styles.xml`. The available paragraph/character style IDs are:

`Normale`, `Titolo`, `Sottotitolo`, `Titolo1`, `Titolo2`, `Titolo3`, `Titolo4`, `Titolo5`, `Titolo6`, `Titolo7`, `Titolo8`, `Titolo9`, `Citazione`, `Citazioneintensa`, `Enfasiintensa`, `Riferimentointenso`, `Paragrafoelenco` (plus character variants like `Titolo1Carattere` … `TitoloCarattere`, `SottotitoloCarattere`, `CitazioneCarattere`, `CitazioneintensaCarattere`).

`Paragraph.SetStyle(name)` and `Run.SetStyle(name)` set the style ID verbatim. An ID not present in the document's `styles.xml` produces a non-fatal warning via `doc.Warnings()` (the file still opens in Word). The examples in this repo use English-named IDs (`Title`, `Heading1`, `Heading2`, `Heading3`, `Subtitle`) for clarity — these emit warnings against the default blank-doc style table but still open. To use English-named styles without warnings, supply a template via `FromTemplate`/`OpenTemplate` whose `styles.xml` defines them.

```go
doc.AddParagraph("Title").SetStyle("Titolo")
doc.AddParagraph("Section").SetStyle("Titolo1")
para.AddRun("important").SetStyle("Enfasiintensa")
```

### Tables

```go
// Simple grid from a string matrix:
tb, _ := doc.AddTable([][]string{{"A","B"},{"1","2"}})

// Full fluent builder:
tbl := doc.AddTableBuilder()
tbl.SetTableStyle("LightGridAccent1")
tbl.SetWidth(8000, "dxa")                       // dxa | pct | auto
tbl.SetBorders(&wordingo.TableBorders{
    Top:    &wordingo.BorderDef{Style:"single", Size:8, Color:"2E75B6"},
    Bottom: &wordingo.BorderDef{Style:"single", Size:8, Color:"2E75B6"},
})
tbl.SetShading("clear", "D9E2F3")
tbl.Row(0).Cell(0).SetText("Product").SetBold(true).SetWidth(1500, "dxa")
tbl.Row(0).Cell(0).SetShading("clear", "D9E2F3")
tbl.Row(0).Cell(0).MergeRight()                 // GridSpan = 2
tbl.Row(0).Cell(0).MergeDown()                 // vMerge restart
if err := tbl.DeleteRow(2); err != nil { /* out of range */ }
for _, t := range doc.Tables() { /* *TableBuilder per body table */ }
```

Tables are appended after all paragraphs (v1 limitation — D-24). Per-cell `Row(idx).Cell(idx)` builders grow the row/cell slices lazily.

### Images

```go
run, err := doc.AddImage("photo.png")           // PNG or JPEG from disk
run, err := doc.AddImageBytes("gradient.png", pngBytes, "image/png")
run.SetImageWidth(3.0).SetImageHeight(2.0)        // display size in inches
```

PNG and JPEG are embedded as DrawingML inline images (`w:drawing/w:inline`). DPI is auto-detected from JFIF APP0 / EXIF / PNG `pHYs` chunks; when unavailable, falls back to 72 DPI and a 3-inch default width with locked aspect ratio. The returned `*Run` is the run carrying the drawing, so image sizing chains on it.

### Headers & footers

```go
h := doc.AddHeader(wordingo.HeaderDefault)       // or HeaderFirst, HeaderEven
h.AddParagraph("Report — Confidential")
h.Paragraphs()
h.InsertParagraphAt(0, ctP)
h.DeleteParagraphAt(0)
f := doc.AddFooter(wordingo.FooterFirst)        // or FooterDefault, FooterEven
f.AddParagraph("Page ")
```

Header/footer OPC parts are created lazily on first `AddHeader`/`AddFooter` and linked to the section's `sectPr` via `headerReference`/`footerReference`.

### Lists

```go
lb := doc.AddList(false)                         // bulleted; true → ordered
lb.AddItem("Apples", 0)                          // level 0-8 (9 levels of nesting)
lb.AddItem("Kitchen", 0).AddItem("Fridge", 1)    // chainable
doc.AddListFromSlice([]string{"One","Two","Three"}, true)  // convenience
lb := doc.AddNumberingDef("upperRoman", 1)       // custom numFmt + start across 9 levels
```

`AddList`/`AddListFromSlice`/`AddNumberingDef` create abstract+numbering definitions in `word/numbering.xml`, merged with any numbering brought over by a template. Auto-generated `abstractNumId` and `numId` scan existing entries to avoid collisions; `numId=0` is reserved for Word's built-in `ListNumber`.

### Hyperlinks

```go
p := doc.AddParagraph("Visit ")
r := p.AddHyperlink("Go docs", "https://go.dev")
r.SetColor("0563C1").SetUnderline("single")      // chain Run formatting on the returned *Run
```

Each `AddHyperlink` call allocates a fresh external relationship in `word/_rels/document.xml.rels` with `TargetMode="External"`. URIs are not deduplicated.

### Page setup

```go
doc.SetOrientation(wordingo.OrientationPortrait)   // or OrientationLandscape
doc.SetPaperSize(wordingo.PaperLetterW, wordingo.PaperLetterH) // Letter/A4/Legal constants, or custom twips
doc.SetPaperSize(wordingo.PaperA4W, wordingo.PaperA4H)
doc.SetMargins(1440, 1440, 1440, 1440)            // top, right, bottom, left twips
doc.AddPageBreak()                                // append a page-break paragraph
// Same calls on the Section wrapper:
doc.Section().SetOrientation(wordingo.OrientationLandscape)
doc.Section().SetPaperSize(wordingo.PaperA4W, wordingo.PaperA4H)
doc.Section().SetMargins(1440, 1440, 1440, 1440)
```

`SetOrientation` swaps W/H in the current `PgSz` so landscape = W > H. All three `Document` setters return `*Document` for chaining.

### Template merge

```go
doc.Merge(map[string]string{
    "name":  "Alice",
    "item":  "invoice",
    "qty_1": "10",
}, nil)                                            // nil opts = all parts (Body,Tables,Headers,Footers)

doc.Merge(map[string]string{"page_num":"2"}, &wordingo.MergeOpts{
    ScopedParts: wordingo.ScopedParts{Headers: true, Footers: true},
})
```

`Merge` replaces `{{key}}` placeholders inside run text, including placeholders that span multiple runs (split-run safe). `nil` opts defaults to scanning all four scopes. Unused keys surface as warnings via `doc.Warnings()`.

### Edit operations

```go
target := /* *Paragraph found via doc.Paragraphs() or doc.Body() */
doc.InsertBefore(target, "Dear Alice,")           // pointer-identity based; returns new *Paragraph
doc.InsertAfter(target, "Please find details below.")
doc.DeleteParagraph(target)                        // pointer-identity based
r.SetText("new text")                              // overwrite run text
r.ReplaceText("old", "new")                        // strings.ReplaceAll on run text
if err := tbl.DeleteRow(2); err != nil { /* OOB */ } // returns error, never panics

for _, p := range doc.Paragraphs() { /* body paragraphs */ }
for _, t := range doc.Tables()      { /* body tables */ }
for _, el := range doc.Body() {
    switch el.Type {
    case wordingo.ElementParagraph: el.Para /* *Paragraph */
    case wordingo.ElementTable:     el.Table /* *TableBuilder */
    }
}
```

`InsertBefore`/`InsertAfter`/`DeleteParagraph` locate the target by pointer identity (`p.ct == target.ct`); a not-found target records a warning rather than mutating the body.

### Text extraction

```go
text, _ := doc.ExtractText(nil)                    // all parts; "\n" separator
text, _ := doc.ExtractText(&wordingo.ExtractOpts{
    ScopedParts: wordingo.ScopedParts{Body:true, Tables:true, Headers:false, Footers:false},
    Separator:   " | ",
})
```

`nil` opts defaults to all four scopes (`Body`, `Tables`, `Headers`, `Footers`) with a `"\n"` separator. Read-only — does not mutate the document.

### Markdown

```go
md, _ := doc.ToMarkdown(nil)                       // GFM; nil opts = all parts
md, _ := doc.ToMarkdown(&wordingo.ExtractOpts{ScopedParts: ...})

d, _  := wordingo.CreateFromMarkdown("# Hi\n\n- a\n- b\n")   // new Document
d.ImportMarkdown("## Appended\n\nMore content")             // append; parse errors → Warnings, never panic
```

`ToMarkdown` emits ATX headers (`Heading1`..`Heading6` → `#`..`######`), `**bold**` / `*italic*` / `~~strike~~`, pipe tables with separator rows, `[text](url)` from hyperlink relationships, `![alt](data:image/png;base64,...)` image data URIs, ordered/bulleted list markers, and fenced code blocks (paragraphs whose runs use `Consolas`/`Courier`/`Courier New`). `CreateFromMarkdown` and `ImportMarkdown` accept headings, paragraphs, ordered/bulleted lists with nesting, pipe tables, fenced code blocks, and inline bold/italic/code/links/data-URI images.

### Validation & escape hatch

```go
for _, w := range doc.Warnings() { fmt.Println(w) }   // non-fatal issues from package + document layers

pkg  := doc.X()                // *opc.Package — content types, rels, parts
ctP  := para.X()                // *wml.CT_P
ctR  := run.X()                 // *wml.CT_R
tbl  := tableBuilder.X()        // *wml.CT_Tbl
tr   := rowBuilder.X()          // *wml.CT_Tr
tc   := cellBuilder.X()         // *wml.CT_Tc
ctP2 := listBuilder.X()         // *wml.CT_P of the last AddItem
hdr  := header.X()              // *wml.CT_Hdr
ftr  := footer.X()              // *wml.CT_Ftr
sp   := section.X()             // *wml.CT_SectPr
```

`Warnings()` aggregates package-layer warnings (OPC, content types, part sizes) and document-layer warnings (unknown style refs, invalid colors, negative sizes, unused merge keys, markdown parse errors, body serialization errors). The `X()` family returns the underlying WML struct for any field the public API does not expose — the preferred escape hatch over adding more methods.

## Examples

Each example below is a standalone Go `main` package. Run it from its directory with `go run main.go`.

1. [examples/01-blank-doc/](examples/01-blank-doc/README.md) — minimum Hello World
2. [examples/02-text-and-styles/](examples/02-text-and-styles/README.md) — paragraphs, headings, inline run formatting
3. [examples/03-tables/](examples/03-tables/README.md) — simple grid + fluent builder with borders, shading, cell width
4. [examples/04-images/](examples/04-images/README.md) — embedded PNG with explicit size
5. [examples/05-headers-footers/](examples/05-headers-footers/README.md) — header/footer, A4, landscape, margins, page break
6. [examples/06-lists/](examples/06-lists/README.md) — bulleted, ordered, nested outline, `AddListFromSlice`
7. [examples/07-hyperlinks/](examples/07-hyperlinks/README.md) — inline links with chained `Run` formatting
8. [examples/08-comprehensive/](examples/08-comprehensive/README.md) — every feature in one doc
9. [examples/09-merge-and-edit/](examples/09-merge-and-edit/README.md) — `{{placeholder}}` Merge, `InsertBefore`/`InsertAfter`, `DeleteParagraph`, `DeleteRow`, `ReplaceText`/`SetText`, scoped Merge, `Body()` iteration, `Warnings()`
10. [examples/10-template-to-document/](examples/10-template-to-document/README.md) — `FromTemplate` (fresh body), `OpenTemplate` (preserved body), `FromTemplate`+`Merge`+`InsertBefore` combined
11. [examples/11-text-extraction-markdown/](examples/11-text-extraction-markdown/README.md) — `ExtractText`, `ToMarkdown`, `CreateFromMarkdown`, `ImportMarkdown`, scoped parts, round-trip verification

## Design & philosophy

- **Pure Go stdlib only.** Zero external dependencies; `go.mod` has no `require` directives. Only the Go standard library (`archive/zip`, `bytes`, `encoding/base64`, `encoding/xml`, `fmt`, `image/jpeg`, `image/png`, `io`, `math`, `os`, `path`, `regexp`, `strings`) plus this module's own internal packages.
- **ISO/IEC 29500 OOXML is the authoritative spec.** Around sixty essential WML types are modeled as Go structs; unknown XML elements are preserved verbatim via raw-element hoisting, never lossy-parsed.
- **Style fidelity is the product.** Templates are cloned by copying the style dependency graph byte-for-byte (styles, numbering, fontTable, theme, settings), never re-encoded. `FromTemplate` always targets a fresh empty body to avoid mutating a user's source file.
- **Open-in-Word-without-repair is the validation bar.** Every producer document is round-tripped through the library and re-opened in Microsoft Word as the acceptance test.
- **Escape hatch over public-surface bloat.** When the structured API lacks a field, the `X()` methods hand back the underlying WML struct for direct mutation — adding a wrapper for every WML element is explicitly out of scope.

## Status

v1.0 milestone complete (Phases 1-6 + 06.1 all shipped). The v1 surface covers document create/open/template, paragraphs/runs, named styles, rich inline formatting, tables (grid + builder), PNG/JPEG images, headers/footers, bulleted/ordered/nested lists, hyperlinks, page setup, template merge, edit operations, body-ordered iteration, and text/markdown round-tripping.

Out of scope for v1 (and not implemented): PDF conversion; `.doc`/`.odt`/`.rtf`/`.xlsx`/`.pptx`; a CLI tool; watch mode; HTML rendering; Word COM automation/CGO; field codes; comments; bookmarks; content controls (SDT); tracked changes; charts; equations. These are deferred to v2.

## License

MIT, copyright 2026 Fabio Marini. See [LICENSE](LICENSE).

## Contributing

PRs welcome. For non-trivial changes, please open an issue first to discuss the approach. Keep the zero-dependency invariant — only the Go standard library and this module's own `internal/` packages are allowed. Run `go test ./...` before submitting.