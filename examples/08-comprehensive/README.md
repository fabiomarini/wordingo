# Example 08: Comprehensive

## What it demonstrates

Every major v1 feature in a single document: header/footer, paper size, `Title`/`Subtitle`/`Heading1`-style paragraphs, inline run formatting, simple grid + fluent builder tables with shading, an embedded PNG image, bulleted + ordered lists, an inline hyperlink, a page break, and a landscape-configured section.

## How to run

```
cd examples/08-comprehensive
go run main.go
```

## Output

- `output.docx` — a multi-page document exercising all surfaces listed above on a single section whose page-size configuration switches to landscape with 0.5-inch margins partway through.

## Code walkthrough

- `doc.AddHeader(wordingo.HeaderDefault).AddParagraph("Wordingo Demo Document — Header")` and `doc.AddFooter(wordingo.FooterDefault).AddParagraph("Page ")` — header/footer OPC parts created lazily on first call, then populated.
- `doc.SetPaperSize(wordingo.PaperLetterW, wordingo.PaperLetterH)` — explicit Letter paper size; this matches the default but pins it.
- `doc.AddParagraph("Wordingo: Pure Go Word Documents").SetStyle("Title")` (and `Subtitle`, `Heading1`, `Heading2`) — named style IDs matching the default blank document's style table.
- `p := doc.AddParagraph("Inline formatting: "); p.AddRun("bold").SetBold(true); p.AddRun(", "); p.AddRun("italic").SetItalic(true); ...` — multiple runs with chained formatting in one paragraph (`SetBold`, `SetItalic`, `SetUnderline`, `SetColor`, `SetFont`, `SetSize`).
- `tbl := doc.AddTableBuilder(); tbl.SetTableStyle("LightGridAccent1"); tbl.SetWidth(8000, "dxa"); tbl.SetBorders(&wordingo.TableBorders{...})` then `tbl.Row(0).Cell(0).SetText("Item").SetBold(true).SetWidth(2500, "dxa").SetShading("", "D9E2F3")` — builder table with table-level style/width/borders and per-cell bold/width/shading.
- `run, err := doc.AddImageBytes("gradient.png", imgData, "image/png"); run.SetImageWidth(2.5).SetImageHeight(2.5)` — 100×100 gradient PNG embedded as DrawingML inline, sized 2.5×2.5 inches via the returned `*Run`.
- `features := doc.AddList(false); features.AddItem("Pure Go — zero external dependencies", 0); ...` — bulleted list; `plan := doc.AddList(true); plan.AddItem("Foundation: ...", 0)` — ordered list.
- `link := doc.AddParagraph("Learn more at "); link.AddHyperlink("GitHub Repository", "https://github.com/fabiomarini/wordingo").SetColor("0563C1").SetUnderline("single")` — external hyperlink with run formatting.
- `doc.AddParagraph("").SetPageBreakBefore(true)` — page break before the next content.
- `doc.SetOrientation(wordingo.OrientationLandscape); doc.SetMargins(720, 720, 720, 720)` — switch the single section to landscape with 0.5-inch margins. These mutate the existing `sectPr`, so the whole document's section flips orientation.