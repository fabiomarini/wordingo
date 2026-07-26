# Templates and merge

## Templates

Templates copy a source .docx's formatting — styles, numbering, font table, theme, settings — into a new document. Two entry points, one for empty bodies and one for preserved bodies.

### FromTemplate — fresh body, template styles

```go
doc, err := wordingo.FromTemplate("corporate-template.docx")
// doc has an empty body with the template's styles
doc.AddParagraph("Quarterly Report").SetStyle("Title")
doc.Save("report.docx")
```

Use this when you want the template's look but write all content yourself.

### OpenTemplate — keep template content

```go
doc, err := wordingo.OpenTemplate("letterhead.docx")
// doc preserves the template's paragraphs and tables
for _, p := range doc.Paragraphs() {
    fmt.Println(p.Text())
}
doc.AddParagraph("Additional content after template text")
doc.Save("output.docx")
```

Use this when the template already has content you want to extend.

### Reader variants

```go
doc, _ := wordingo.FromTemplateReader(r, size)
doc, _ := wordingo.OpenTemplateReader(r, size)
```

Both accept `io.ReaderAt` for situations where you don't have a file path.

### How it works

Style parts (`styles.xml`, `numbering.xml`, `fontTable.xml`, `theme1.xml`, `settings.xml`) are copied byte-for-byte — never parsed or re-encoded. This guarantees the cloned formatting is identical to the source.

Header and footer relationship IDs in the section properties are stripped (they would point to parts not present in the clone).

## Template merge

Replace `{{placeholder}}` text in body, tables, headers, and footers.

```go
doc.Merge(map[string]string{
    "name":    "Alice",
    "item":    "widget",
    "qty_1":   "10",
    "price_1": "$5.00",
}, nil)   // nil opts = scan all parts
```

Scoped merge — only certain parts:

```go
doc.Merge(map[string]string{
    "page_num": "2",
}, &wordingo.MergeOpts{
    ScopedParts: wordingo.ScopedParts{
        Headers: true,
        Footers: true,
    },
})
```

### Split-run handling

Placeholders that span multiple formatting runs (e.g., `{{` in one run and `name` in the next) are detected and rewritten into a single text value. Non-text runs (line breaks, tabs, images) are preserved.

### Missing keys

Keys from the merge map not found anywhere in the document surface as warnings:

```go
for _, w := range doc.Warnings() {
    fmt.Println(w)   // "merge key 'page_num' not found in document"
}
```

## MergeOpts

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ScopedParts` | `ScopedParts` | all true | Which document parts to scan for placeholders |

## ScopedParts

| Field | Type | Description |
|-------|------|-------------|
| `Body` | `bool` | Body paragraphs |
| `Tables` | `bool` | Table cells |
| `Headers` | `bool` | Header paragraphs |
| `Footers` | `bool` | Footer paragraphs |

## Edit operations

### Paragraph editing

All editing targets are identified by pointer identity — the `*Paragraph` you pass must be one returned by `AddParagraph`, `Paragraphs()`, or `Body()`.

```go
target := doc.Paragraphs()[2]

inserted := doc.InsertBefore(target, "New text before")
inserted := doc.InsertAfter(target,  "New text after")
doc.DeleteParagraph(target)
```

If the target paragraph is not found in the document body, a warning is recorded.

### Run text editing

```go
r.SetText("new text")             // overwrite run content
r.ReplaceText("old", "new")       // strings.ReplaceAll on run content
```

### Table row editing

```go
if err := tbl.DeleteRow(2); err != nil {
    // out of range — returns error, never panics
}
```
