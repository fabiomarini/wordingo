# Document structure: headers, footers, page setup, sections

## Headers and footers

Three variants per section: default, first-page, and even-page.

```go
// Add headers
h := doc.AddHeader(wordingo.HeaderDefault)
h.AddParagraph("Confidential — Quarterly Report")

h2 := doc.AddHeader(wordingo.HeaderFirst)    // first page only
h3 := doc.AddHeader(wordingo.HeaderEven)     // even pages

// Add footers
f := doc.AddFooter(wordingo.FooterDefault)
f.AddParagraph("Page ")
```

Once created, headers and footers support full paragraph editing:

```go
for _, p := range h.Paragraphs() {
    fmt.Println(p.Text())
}

// Insert at a specific position
h.InsertParagraphAt(0, ctP)
h.DeleteParagraphAt(0)
```

Headers and footers are created lazily — the OPC part is only written when `AddHeader` or `AddFooter` is first called.

## Page setup

### Orientation

```go
doc.SetOrientation(wordingo.OrientationPortrait)     // default
doc.SetOrientation(wordingo.OrientationLandscape)
```

Swaps width and height in the section properties so landscape renders as W > H.

### Paper size

```go
doc.SetPaperSize(wordingo.PaperLetterW, wordingo.PaperLetterH)   // Letter (default)
doc.SetPaperSize(wordingo.PaperA4W, wordingo.PaperA4H)           // A4
doc.SetPaperSize(wordingo.PaperLegalW, wordingo.PaperLegalH)     // Legal
```

Custom sizes in twips (1/1440 inch):

```go
doc.SetPaperSize(12240, 15840)
```

### Margins

```go
// top, right, bottom, left in twips
doc.SetMargins(1440, 1440, 1440, 1440)   // 1 inch all sides
```

All three setters (`SetOrientation`, `SetPaperSize`, `SetMargins`) return `*Document` for chaining.

### Page breaks

```go
// As a separate paragraph
doc.AddPageBreak()

// Or on an existing paragraph
p.SetPageBreakBefore(true)
```

### Section wrapper

The same page setup methods are also available on a `Section` wrapper:

```go
s := doc.Section()
s.SetOrientation(wordingo.OrientationLandscape)
s.SetPaperSize(wordingo.PaperA4W, wordingo.PaperA4H)
s.SetMargins(720, 720, 720, 720)
```

`Section.X()` returns the underlying WML section properties struct for low-level access.

## Body iteration

`Body()` returns paragraphs and tables in document order — not just paragraphs first, then tables:

```go
for _, el := range doc.Body() {
    switch el.Type {
    case wordingo.ElementParagraph:
        // el.Para is *Paragraph
        fmt.Println(el.Para.Text())
    case wordingo.ElementTable:
        // el.Table is *TableBuilder
        for _, row := range el.Table.Rows() { ... }
    }
}
```

Backward compatible: `Paragraphs()` and `Tables()` still work as before.
