# API reference

## Package `wordingo`

Import path: `github.com/fabiomarini/wordingo`

---

## Document

Main document type. Created via `Create`, `Open`, `OpenReader`, `FromTemplate`, `OpenTemplate`, or `CreateFromMarkdown`.

### Document methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Save` | `(path string) error` | Write document to file |
| `WriteTo` | `(w io.Writer) (int64, error)` | Write to any writer |
| `SaveFile` | `(path string) error` | Alternate file writer |
| `Close` | `() error` | Release internal references |
| `Warnings` | `() []string` | Non-fatal issues |
| `X` | `() *opc.Package` | Underlying OPC package |
| `Paragraphs` | `() []*Paragraph` | Body paragraphs |
| `Tables` | `() []*TableBuilder` | Body tables |
| `Body` | `() []BodyElement` | Paragraphs + tables in document order |
| `AddParagraph` | `(text string) *Paragraph` | Append a paragraph |
| `InsertBefore` | `(target *Paragraph, text string) *Paragraph` | Insert before target |
| `InsertAfter` | `(target *Paragraph, text string) *Paragraph` | Insert after target |
| `DeleteParagraph` | `(target *Paragraph)` | Remove a paragraph |
| `AddTable` | `(data [][]string) (*TableBuilder, error)` | Quick table from grid |
| `AddTableBuilder` | `() *TableBuilder` | Fluent table builder |
| `AddImage` | `(path string) (*Run, error)` | Embed image from file |
| `AddImageBytes` | `(name string, data []byte, contentType string) (*Run, error)` | Embed image from bytes |
| `AddHeader` | `(variant HeaderVariant) *Header` | Add a header |
| `AddFooter` | `(variant FooterVariant) *Footer` | Add a footer |
| `AddList` | `(ordered bool) *ListBuilder` | Start a list |
| `AddListFromSlice` | `(items []string, ordered bool)` | Convenience flat list |
| `AddNumberingDef` | `(numFmt string, start int64) *ListBuilder` | Custom numbering format |
| `AddPageBreak` | `() *Document` | Append page-break paragraph |
| `SetOrientation` | `(o PageOrientation) *Document` | Page orientation |
| `SetPaperSize` | `(w, h int64) *Document` | Paper dimensions in twips |
| `SetMargins` | `(top, right, bottom, left int64) *Document` | Margins in twips |
| `Section` | `() *Section` | Section properties wrapper |
| `Merge` | `(data map[string]string, opts *MergeOpts) error` | Replace `{{placeholders}}` |
| `ExtractText` | `(opts *ExtractOpts) (string, error)` | Plain text extraction |
| `ToMarkdown` | `(opts *ExtractOpts) (string, error)` | Markdown export |
| `ImportMarkdown` | `(md string)` | Append markdown content |
| `Headings` | `() []Heading` | Document outline (Heading1–9 styles / outline levels) |
| `AddTableOfContents` | `(opts *TOCOptions) (*TOC, error)` | Append a TOC section built from the headings |
| `InsertTableOfContentsBefore` | `(target *Paragraph, opts *TOCOptions) (*TOC, error)` | Insert a TOC section before a paragraph |

---

## Constructors

| Function | Signature | Description |
|----------|-----------|-------------|
| `Create` | `() (*Document, error)` | Blank document |
| `Open` | `(path string) (*Document, error)` | Open existing .docx |
| `OpenReader` | `(r io.ReaderAt, size int64) (*Document, error)` | Open from reader |
| `FromTemplate` | `(path string) (*Document, error)` | Template → fresh body |
| `OpenTemplate` | `(path string) (*Document, error)` | Template → preserved body |
| `FromTemplateReader` | `(r io.ReaderAt, size int64) (*Document, error)` | Template from reader |
| `OpenTemplateReader` | `(r io.ReaderAt, size int64) (*Document, error)` | Template from reader |
| `CreateFromMarkdown` | `(md string) (*Document, error)` | Document from markdown |

---

## Paragraph

| Method | Signature | Description |
|--------|-----------|-------------|
| `Text` | `() string` | Concatenated run text |
| `Style` | `() string` | Paragraph style ID |
| `AddRun` | `(text string) *Run` | Append a run |
| `SetStyle` | `(name string) *Paragraph` | Paragraph style |
| `SetAlignment` | `(a Alignment) *Paragraph` | Text alignment |
| `SetSpacing` | `(s *ParSpacing) *Paragraph` | Paragraph spacing |
| `SetIndent` | `(i *ParIndent) *Paragraph` | Paragraph indent |
| `SetPageBreakBefore` | `(b bool) *Paragraph` | Page break before |
| `SetFormatting` | `(f ParFormat) *Paragraph` | Bulk set formatting |
| `X` | `() *wml.CT_P` | Raw WML paragraph |

---

## Run

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetBold` | `(b bool) *Run` | Bold |
| `SetItalic` | `(b bool) *Run` | Italic |
| `SetUnderline` | `(s string) *Run` | Underline style |
| `SetFont` | `(s string) *Run` | Font family |
| `SetSize` | `(pts float64) *Run` | Font size in points |
| `SetColor` | `(s string) *Run` | Hex color (6 digits) |
| `SetHighlight` | `(s string) *Run` | Highlight color |
| `SetStyle` | `(s string) *Run` | Character style |
| `SetText` | `(s string) *Run` | Overwrite run text |
| `ReplaceText` | `(old, new string) *Run` | Replace in run text |
| `SetFormatting` | `(f RunFormat) *Run` | Bulk set formatting |
| `AddHyperlink` | `(text, url string) *Run` | Clickable link |
| `SetImageWidth` | `(inches float64) *Run` | Image display width |
| `SetImageHeight` | `(inches float64) *Run` | Image display height |
| `X` | `() *wml.CT_R` | Raw WML run |

---

## TableBuilder

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetTableStyle` | `(s string) *TableBuilder` | Table style |
| `SetWidth` | `(w int64, t string)` | Table width + type |
| `SetBorders` | `(b *TableBorders) *TableBuilder` | Table borders |
| `SetShading` | `(s, c string) *TableBuilder` | Table shading |
| `DeleteRow` | `(idx int) error` | Delete row by index |
| `Row` | `(idx int) *RowBuilder` | Access row (grows lazily) |
| `X` | `() *wml.CT_Tbl` | Raw WML table |

## RowBuilder

| Method | Signature | Description |
|--------|-----------|-------------|
| `Cell` | `(idx int) *CellBuilder` | Access cell (grows lazily) |
| `SetBorders` | `(b *TableBorders) *RowBuilder` | Row-level borders |
| `X` | `() *wml.CT_Tr` | Raw WML row |

## CellBuilder

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetText` | `(s string) *CellBuilder` | Cell text |
| `SetBold` | `(b bool) *CellBuilder` | Bold text in cell |
| `SetShading` | `(s, c string) *CellBuilder` | Cell background |
| `SetWidth` | `(w int64, t string) *CellBuilder` | Cell width |
| `MergeRight` | `() *CellBuilder` | Merge with right cell |
| `MergeDown` | `() *CellBuilder` | Merge with cell below |
| `X` | `() *wml.CT_Tc` | Raw WML cell |

---

## ListBuilder

| Method | Signature | Description |
|--------|-----------|-------------|
| `AddItem` | `(text string, level int) *ListBuilder` | Add item at level (0–8) |
| `X` | `() *wml.CT_P` | Raw WML paragraph of last item |

---

## Header / Footer

Both implement `ParagraphContainer`.

| Method | Signature | Description |
|--------|-----------|-------------|
| `AddParagraph` | `(text string) *Paragraph` | Append paragraph |
| `Paragraphs` | `() []*Paragraph` | All paragraphs |
| `InsertParagraphAt` | `(idx int, p *wml.CT_P)` | Insert at position |
| `DeleteParagraphAt` | `(idx int)` | Delete at position |
| `X` | `() *wml.CT_Hdr` / `*wml.CT_Ftr` | Raw WML struct |

---

## Section

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetOrientation` | `(o PageOrientation) *Section` | Page orientation |
| `SetPaperSize` | `(w, h int64) *Section` | Paper size |
| `SetMargins` | `(top, right, bottom, left int64) *Section` | Margins |
| `X` | `() *wml.CT_SectPr` | Raw WML section |

---

## Types

```go
type Alignment int

type RunFormat struct {
    Bold      *bool
    Italic    *bool
    Underline *string
    Font      *string
    Size      *float64
    Color     *string
    Highlight *string
}

type ParFormat struct {
    Alignment *Alignment
    Spacing   *ParSpacing
    Indent    *ParIndent
}

type ParSpacing struct {
    Before, After, Line int64
    LineRule            string
}

type ParIndent struct {
    Left, Right, FirstLine, Hanging int64
}

type TableBorders struct {
    Top, Bottom, Left, Right, InsideH, InsideV *BorderDef
}

type BorderDef struct {
    Style string
    Size  int64
    Color string
}

type HeaderVariant int   // HeaderDefault, HeaderFirst, HeaderEven
type FooterVariant int   // FooterDefault, FooterFirst, FooterEven
type PageOrientation int // OrientationPortrait, OrientationLandscape

type ScopedParts struct {
    Body, Tables, Headers, Footers bool
}

type MergeOpts struct {
    ScopedParts ScopedParts
}

type ExtractOpts struct {
    ScopedParts ScopedParts
    Separator   string
}

type Heading struct {
    Level int    // 1..9
    Text  string // trimmed heading text
    Style string // style id, e.g. "Heading1" ("" if from outline level)
}

type TOCOptions struct {
    Title        string // default "Table of Contents"
    Levels       int    // deepest heading level (1..9), default 3
    UpdateOnOpen *bool  // nil/true: set w:updateFields so Word refreshes on open
}

type TOC struct {
    // returned by AddTableOfContents / InsertTableOfContentsBefore
}

func (t *TOC) Title() *Paragraph // the TOC title paragraph

type BodyElementType int   // ElementParagraph, ElementTable

type BodyElement struct {
    Type  BodyElementType
    Para  *Paragraph
    Table *TableBuilder
}
```

---

## Constants

### Alignment

```go
AlignmentLeft
AlignmentCenter
AlignmentRight
AlignmentBoth
```

### Header/Footer variants

```go
HeaderDefault
HeaderFirst
HeaderEven

FooterDefault
FooterFirst
FooterEven
```

### Page orientation

```go
OrientationPortrait
OrientationLandscape
```

### Paper sizes (twips)

```go
PaperLetterW = 12240    // Letter width
PaperLetterH = 15840    // Letter height
PaperA4W     = 11906    // A4 width
PaperA4H     = 16838    // A4 height
PaperLegalW  = 12240    // Legal width
PaperLegalH  = 20160    // Legal height
```

### Body element types

```go
ElementParagraph BodyElementType = iota
ElementTable
```
