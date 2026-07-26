# Example 04: Images

## What it demonstrates

Embedding a PNG into a document as a DrawingML inline image via `AddImageBytes`, with explicit display sizing in inches via the returned `*Run`. Also shows how to manufacture a PNG with the standard library `image/png` encoder.

## How to run

```
cd examples/04-images
go run main.go
```

## Output

- `output.docx` — a single-page document containing a 100×100 gradient PNG sized to 3×2 inches.

## Code walkthrough

- `imgData := makeGradientPNG()` — helper builds a 100×100 `image.NRGBA`, encodes it to PNG via `image/png.Encode` from the standard library, returns the raw bytes. The actual image API in wordingo accepts any PNG/JPEG byte slice.
- `run, err := doc.AddImageBytes("gradient.png", imgData, "image/png")` — embeds the bytes as `word/media/imageN.png`, registers a content-type override, and adds a fresh relationship from `word/document.xml` to the media part. The image is appended to the last paragraph in the body as a DrawingML inline element; returns `(*Run, error)`.
- `run.SetImageWidth(3.0).SetImageHeight(2.0)` — mutates the `w:extent` (EMU) on the returned run's inline drawing; values are inches, converted to EMU via `inches * 914400`. Aspect ratio is not auto-locked when both setters are called explicitly.
- DPI detection: when a PNG `pHYs` or JPEG JFIF/EXIF chunk is present, wordingo recomputes the EMU extent from pixel dimensions and detected DPI, capped at a 3-inch default width with proportional height. When DPI is unknown, falls back to 72 DPI and the 3-inch default.