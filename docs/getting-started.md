# Getting started

## Installation

```bash
go get github.com/fabiomarini/wordingo
```

Requires Go 1.23+.

## Quick start

Create a blank document, add content, and save:

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

## Create, open, save

```go
// Blank document with default styles (Letter, 1-inch margins)
doc, _ := wordingo.Create()
defer doc.Close()

// Save to file
doc.Save("output.docx")

// Write to any io.Writer
n, _ := doc.WriteTo(w)   // returns bytes written

// Alternate file writer
doc.SaveFile("output.docx")
```

```go
// Open existing .docx for editing
doc, _ := wordingo.Open("existing.docx")
defer doc.Close()

// From io.ReaderAt
doc, _ := wordingo.OpenReader(r, size)
```

All changes are tracked internally and serialized at Save time.

## Warnings

Non-fatal issues surface via `Warnings()`:

```go
for _, w := range doc.Warnings() {
    fmt.Println(w)
}
```

Examples: unknown style names, invalid color hex values, unused merge keys, markdown parse errors. The document still opens in Word — warnings are informational.
