# Tables and images

## Tables

Two styles: a quick grid from a string matrix, or a full fluent builder.

### Quick grid

```go
tbl, _ := doc.AddTable([][]string{
    {"Item", "Price", "Qty"},
    {"Widget", "$5.00", "10"},
    {"Gadget", "$25.00", "3"},
})
```

Returns `(*TableBuilder, error)`. Errors on empty or jagged data.

### Fluent builder

```go
tbl := doc.AddTableBuilder()
tbl.SetTableStyle("LightGridAccent1")
tbl.SetWidth(8000, "dxa")      // dxa | pct | auto
tbl.SetBorders(&wordingo.TableBorders{
    Top:    &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
    Bottom: &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
    Left:   &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
    Right:  &wordingo.BorderDef{Style: "single", Size: 8, Color: "2E75B6"},
})
tbl.SetShading("clear", "D9E2F3")
```

### Cell operations

Rows and cells are accessed by index. Slices grow lazily.

```go
tbl.Row(0).Cell(0).SetText("Header").SetBold(true)
tbl.Row(0).Cell(0).SetShading("clear", "D9E2F3")
tbl.Row(0).Cell(0).SetWidth(2500, "dxa")

// Merge cells
tbl.Row(0).Cell(0).MergeRight()   // GridSpan = 2
tbl.Row(0).Cell(0).MergeDown()    // vMerge restart

// Row-level borders
tbl.Row(0).SetBorders(&wordingo.TableBorders{
    Bottom: &wordingo.BorderDef{Style: "single", Size: 6, Color: "2E75B6"},
})

// Delete a row
if err := tbl.DeleteRow(2); err != nil { /* out of range */ }
```

### Iterating tables

```go
for _, t := range doc.Tables() { /* *TableBuilder */ }
```

Tables are appended after all paragraphs in the document body (v1 behavior). Use `Body()` to iterate paragraphs and tables in document order.

## Images

```go
// From disk
run, err := doc.AddImage("photo.png")

// From bytes
run, err := doc.AddImageBytes("gradient.png", pngData, "image/png")
```

PNG and JPEG are both supported. Images are embedded as DrawingML inline.

### Sizing

```go
run.SetImageWidth(3.0).SetImageHeight(2.0)   // display size in inches
```

DPI is auto-detected:
- JPEG: JFIF APP0 and EXIF IFD tags
- PNG: pHYs chunk

When DPI is unavailable, the library falls back to 72 DPI and a 3-inch default width with locked aspect ratio.

### Return value

`AddImage` and `AddImageBytes` return `*Run` — the run that carries the DrawingML element. Image sizing chains off the run so `SetImageWidth` and `SetImageHeight` are called on the returned value directly.
