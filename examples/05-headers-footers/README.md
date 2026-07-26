# Example 05: Headers and Footers

## What it demonstrates

Header/footer part creation, page-size / orientation / margin configuration, and forcing a page break via `SetPageBreakBefore`. Combines the page-setup and header/footer surfaces onto a single section.

## How to run

```
cd examples/05-headers-footers
go run main.go
```

## Output

- `output.docx` — an A4 landscape document with a default header ("Report — Confidential") and a default footer ("Page "), plus a page 2 forced via `SetPageBreakBefore`.

## Code walkthrough

- `doc.AddHeader(wordingo.HeaderDefault).AddParagraph("Report — Confidential")` — creates a header OPC part (`word/headerN.xml`), links it to the section's `sectPr` via a `headerReference` with `type="default"`, and chains `Header.AddParagraph` to populate it. `HeaderDefault` / `HeaderFirst` / `HeaderEven` are the three variants; `AddFooter`/`FooterDefault`/`FooterFirst`/`FooterEven` mirror them.
- `doc.SetPaperSize(wordingo.PaperA4W, wordingo.PaperA4H)` — sets `w:pgSz` to A4 (11906×16838 twips). Constants `PaperLetterW/H`, `PaperA4W/H`, `PaperLegalW/H` are available; custom twips are accepted too. This and the next two `Document` setters return `*Document` for chaining.
- `doc.SetMargins(1440, 1440, 1440, 1440)` — sets `w:pgMar` margins in twips in `(top, right, bottom, left)` order; 1440 twips = 1 inch.
- `doc.SetOrientation(wordingo.OrientationLandscape)` — swaps `w:pgSz` W and H so W > H (landscape). `OrientationPortrait` restores W < H.
- `doc.AddParagraph("").SetPageBreakBefore(true)` — sets `w:pageBreakBefore` on the paragraph so the paragraph renders at the top of a new page. `Document.AddPageBreak` is the shortcut for appending an empty page-break paragraph.