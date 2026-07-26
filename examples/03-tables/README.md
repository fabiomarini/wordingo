# Example 03: Tables

## What it demonstrates

Both ways of building a table: a simple grid from a `[][]string` matrix (`AddTable`) and a full fluent builder (`AddTableBuilder`) configured with a table style, borders, per-cell bold/width — plus the chained `Row().Cell()` builder API.

## How to run

```
cd examples/03-tables
go run main.go
```

## Output

- `output.docx` — two tables: a simple 4×3 grid, then a 4×3 builder table styled with `LightGridAccent1` and blue top/bottom borders.

## Code walkthrough

- `doc.AddTable([][]string{{"Name","Language","Paradigm"}, ...})` — builds a grid table from a string matrix. The returned `*TableBuilder` is ignored here; each row gets one text run per cell. Returns an error if `data` is empty or has no columns.
- `tbl := doc.AddTableBuilder()` — appends an empty `w:tbl` to the body (tables are appended after all paragraphs — v1 limitation, D-24) and returns a builder for fuller control.
- `tbl.SetTableStyle("LightGridAccent1")` — sets the table-style reference; like paragraph styles, unknown names still open in Word.
- `tbl.SetBorders(&wordingo.TableBorders{Top: &wordingo.BorderDef{Style:"single", Size:8, Color:"2E75B6"}, Bottom: ...})` — configures table-level borders via the `TableBorders` / `BorderDef` value types.
- `tbl.Row(0).Cell(0).SetText("Product").SetBold(true)` — chained per-cell builder. `Row(idx)` and `Cell(idx)` grow the underlying row/cell slices lazily (so `Row(3)` on a new table just works). `SetBold` toggles the first run of the first paragraph in the cell; `SetWidth(1500, "dxa")` sets the cell width in twips (`dxa` / `pct` / `auto`).