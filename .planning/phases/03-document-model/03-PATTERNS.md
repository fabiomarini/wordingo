# Phase 3: Document Model - Pattern Map

**Mapped:** 2026-07-25
**Files analyzed:** 10 (4 new, 3 modify, 1 new dir, 2 refactored tests)
**Analogs found:** 8 / 8 (2 N/A for testdata dir and doc.go)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `wordingo/open.go` | orchestrator | CRUD | `internal/opc/package.go:Open()` | role-match |
| `wordingo/template.go` | orchestrator | CRUD | `internal/style/cloner.go:CloneStyles()` | role-match |
| `wordingo/paragraph.go` | model | read-only accessor | `wordingo/wordingo.go:Document` (wrapper pattern) | role-match |
| `wordingo/wordingo.go` | controller | CRUD | `wordingo/wordingo.go` (self — modify existing) | exact |
| `wordingo/create.go` | utility | CRUD | `wordingo/create.go` (self — add `buildEmptyBodyXML`) | exact |
| `wordingo/doc.go` | utility | documentation | N/A (package doc, no code analog) | N/A |
| `wordingo/roundtrip_test.go` | test | CRUD | `internal/opc/opc_test.go:DiffParts/TestRoundTrip` | exact |
| `wordingo/template_test.go` | test | CRUD | `internal/style/cloner_test.go` | exact |
| `wordingo/create_test.go` | test | CRUD | `wordingo/create_test.go` (self — update Save→WriteTo) | exact |
| `testdata/roundtrip/` | fixture | testdata | `testdata/word/` (existing .docx fixtures) | role-match |

## Pattern Assignments

### `wordingo/open.go` (orchestrator, CRUD)

**Analog:** `internal/opc/package.go` lines 139-221 (`Open` function)

**Imports pattern** (package.go lines 1-17):
```go
package opc

import (
    "archive/zip"
    "bytes"
    "fmt"
    "io"
    "strings"
    "unicode/utf8"
)
```

For `wordingo/open.go`, replace with:
```go
package wordingo

import (
    "bytes"
    "encoding/xml"
    "fmt"
    "io"
    "os"

    "github.com/fabiomarini/wordingo/internal/opc"
    "github.com/fabiomarini/wordingo/internal/wml"
    "github.com/fabiomarini/wordingo/internal/xmlutil"
)
```

**Core Open pattern** (package.go:135-221 — Open with eager parse of critical parts, lazy for payloads):
```go
// Open reads an OPC package from r, enforcing all OPC-07 safety limits
// before parsing: entry-count cap, part-name validation, decompressed
// size caps, and compression-ratio flagging. [Content_Types].xml and
// all .rels parts are parsed eagerly; part payloads stay lazy.
func Open(r io.ReaderAt, size int64) (*Package, error) {
    zr, err := zip.NewReader(r, size)
    if err != nil {
        return nil, fmt.Errorf("opc: open zip: %w", err)
    }
    // ... entry enumeration, validation, part creation ...

    // [Content_Types].xml is mandatory — parsed eagerly.
    ctPart, ok := pkg.Parts["[Content_Types].xml"]
    if !ok {
        return nil, fmt.Errorf("opc: missing [Content_Types].xml: %w", ErrInvalidPackage)
    }
    ctBytes, err := readAllCapped(ctPart)
    // ... parse content types ...

    // Parse all .rels parts eagerly.
    for name, part := range pkg.Parts {
        if !isRelsPath(name) { continue }
        // ... read and parse relationships ...
    }

    pkg.Conformance = pkg.detectConformance()
    return pkg, nil
}
```

**Phase 3 adaptation** — `open.go` wraps `opc.Open` + eager body parse (D-03):
```go
func Open(path string) (*Document, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("wordingo: open %s: %w", path, err)
    }
    defer f.Close()
    fi, err := f.Stat()
    if err != nil {
        return nil, fmt.Errorf("wordingo: stat %s: %w", path, err)
    }
    return OpenReader(f, fi.Size())
}

func OpenReader(r io.ReaderAt, size int64) (*Document, error) {
    pkg, err := opc.Open(r, size)
    if err != nil {
        return nil, fmt.Errorf("wordingo: %w", err)
    }

    // Eager parse of word/document.xml (D-03)
    docPart, ok := pkg.Parts["word/document.xml"]
    if !ok {
        return nil, fmt.Errorf("wordingo: missing word/document.xml: %w", opc.ErrInvalidPackage)
    }
    rc, err := docPart.Open()
    if err != nil {
        return nil, fmt.Errorf("wordingo: open document.xml: %w", err)
    }
    defer rc.Close()

    dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
    var doc wml.CT_Document
    if err := dec.Decode(&doc); err != nil {
        return nil, fmt.Errorf("wordingo: decode document.xml: %w", err)
    }

    // Never call MarkModified on style parts (D-06)
    return &Document{pkg: pkg, doc: &doc}, nil
}
```

**Error handling pattern** — wrap with context, use `fmt.Errorf("wordingo: ...: %w", err)` (see package.go:142, 147, 185, 190, 204).

**Auth/Guard pattern:** N/A — no auth for library.

---

### `wordingo/template.go` (orchestrator, CRUD)

**Analog:** `internal/style/cloner.go` lines 88-183 (`CloneStyles`)

**Imports pattern** (cloner.go lines 40-46):
```go
import (
    "fmt"
    "io"
    "strings"

    "github.com/fabiomarini/wordingo/internal/opc"
)
```

For `wordingo/template.go`, use:
```go
package wordingo

import (
    "fmt"
    "io"
    "os"

    "github.com/fabiomarini/wordingo/internal/opc"
    "github.com/fabiomarini/wordingo/internal/style"
    "github.com/fabiomarini/wordingo/internal/wml"
    "github.com/fabiomarini/wordingo/internal/xmlutil"
)
```

**Core clone + body policy pattern** (cloner.go:118-183 — two-pass: precondition scan + byte copy):
```go
func CloneStyles(src, dst *opc.Package) error {
    // Step 1 — Precondition scan (atomicity, D-08).
    for _, p := range cloneParts {
        if _, exists := dst.Parts[p.name]; exists {
            return fmt.Errorf("style: clone %s: %w", p.name, ErrCloneTargetNotEmpty)
        }
    }

    // Step 2 — Byte copy, relationships, content types.
    for _, p := range cloneParts {
        srcPart, ok := src.Parts[p.name]
        if !ok { continue }  // absent source part — skip silently
        rc, err := srcPart.Open()
        // ... read bytes ...
        dst.MarkModified(p.name, bytes)
        dst.ContentTypes.Overrides["/"+p.name] = p.ct
        // ... relationship wiring ...
    }
    return nil
}
```

**Phase 3 adaptation** — `FromTemplate` / `OpenTemplate`:
```go
func FromTemplate(path string) (*Document, error) {
    f, err := os.Open(path)
    if err != nil { return nil, err }
    defer f.Close()
    fi, err := f.Stat()
    if err != nil { return nil, err }
    return FromTemplateReader(f, fi.Size())
}

func FromTemplateReader(r io.ReaderAt, size int64) (*Document, error) {
    src, err := opc.Open(r, size)
    if err != nil {
        return nil, fmt.Errorf("wordingo: open template: %w", err)
    }

    dst := newBlankPackage()
    if err := style.CloneStyles(src, dst); err != nil {
        // cleanup src? src not persisted — just return
        return nil, fmt.Errorf("wordingo: clone styles: %w", err)
    }

    // FromTemplate clears body (CREATE-03)
    emptyBody := buildEmptyBodyXML()
    dst.MarkModified("word/document.xml", emptyBody)

    doc, err := parseDocument(dst)
    if err != nil { return nil, err }
    return &Document{pkg: dst, doc: doc}, nil
}
```

**Error handling pattern** — wrap errors with `fmt.Errorf("wordingo: %s: %w")` (cloner.go:147, cloner.go:152).

---

### `wordingo/paragraph.go` (model, read-only accessor)

**Analog:** `wordingo/wordingo.go` lines 22-57 (`Document` struct + X() escape hatch)

**Imports pattern** (wordingo.go lines 13-18):
```go
import (
    "io"
    "os"

    "github.com/fabiomarini/wordingo/internal/opc"
)
```

For `wordingo/paragraph.go`:
```go
package wordingo

import (
    "strings"

    "github.com/fabiomarini/wordingo/internal/wml"
)
```

**Wrapper-over-schema with X() pattern** (wordingo.go:53-57):
```go
// X returns the underlying OPC package for escape-hatch access to
// internals (content types, relationships, individual parts).
func (d *Document) X() *opc.Package {
    return d.pkg
}
```

**Phase 3 adaptation** — `Paragraph` wrapping `*wml.CT_P` (from CONTEXT.md D-02 and RESEARCH.md Pattern 1):
```go
// Paragraph wraps a WordprocessingML paragraph (w:p).
// Accessors are read-only in Phase 3; Phase 4 adds mutation.
type Paragraph struct {
    ct *wml.CT_P
}

// Style returns the paragraph style ID, or "" if none.
func (p *Paragraph) Style() string {
    if p.ct.PPr != nil && p.ct.PPr.PStyle != nil && p.ct.PPr.PStyle.Val != nil {
        return *p.ct.PPr.PStyle.Val
    }
    return ""
}

// Text returns all text content concatenated across runs.
func (p *Paragraph) Text() string {
    var b strings.Builder
    for _, r := range p.ct.R {
        if r.T != nil {
            b.WriteString(r.T.Value)
        }
    }
    return b.String()
}

// X returns the underlying CT_P for escape-hatch access.
func (p *Paragraph) X() *wml.CT_P { return p.ct }
```

---

### `wordingo/wordingo.go` (controller, CRUD — MODIFY)

**Self-analog:** Existing `wordingo.go` lines 1-57

**Core change: Save → WriteTo rename + new Save(path)** (existing lines 33-46):
```go
// Current (to be renamed):
func (d *Document) Save(w io.Writer) error {
    return d.pkg.Save(w)
}

// SaveFile writes the document to a file at path.
func (d *Document) SaveFile(path string) error {
    f, err := os.Create(path)
    if err != nil { return err }
    defer f.Close()
    return d.pkg.Save(f)
}
```

**Phase 3 replacement** (D-07, per RESEARCH.md Pitfall 5):
```go
// WriteTo writes the document to w. Returns bytes written.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
    n, err := d.pkg.Save(w)  // opc.Save may need to return (int64, error)
    return n, err
}

// Save writes the document to a file at path.
func (d *Document) Save(path string) error {
    f, err := os.Create(path)
    if err != nil { return err }
    defer f.Close()
    _, err = d.WriteTo(f)
    return err
}
```

**Note:** `opc.Package.Save()` currently returns `error`. Phase 3 may need to change its signature to `(int64, error)` or Document wraps it without counting bytes. Check `opc.Package.Save` return type.

---

### `wordingo/create.go` (utility, CRUD — MODIFY)

**Self-analog:** Existing `create.go` lines 1-148

**Core pattern to extend** — add `buildEmptyBodyXML()` (from create.go:103-146 `buildDocumentXML`):
```go
func buildDocumentXML() []byte {
    sz12240 := int64(12240)
    sz15840 := int64(15840)
    margin1440 := int64(1440)
    valNormal := "Normal"

    doc := &wml.CT_Document{
        Body: &wml.CT_Body{
            P: []*wml.CT_P{
                {PPr: &wml.CT_PPr{PStyle: &wml.CT_PStyle{Val: &valNormal}}},
            },
            SectPr: &wml.CT_SectPr{ /* page config */ },
        },
    }
    // ... encode with xmlutil.NewEncoder ...
}
```

**Phase 3 addition** — `buildEmptyBodyXML()` similar but with no paragraphs (for FromTemplate CREATE-03):
```go
func buildEmptyBodyXML() []byte {
    doc := &wml.CT_Document{
        Body: &wml.CT_Body{
            SectPr: &wml.CT_SectPr{
                PgSz: &wml.CT_PgSz{W: ptrInt64(12240), H: ptrInt64(15840)},
                PgMar: &wml.CT_PgMar{
                    Top: ptrInt64(1440), Right: ptrInt64(1440),
                    Bottom: ptrInt64(1440), Left: ptrInt64(1440),
                },
            },
        },
    }
    // ... encode with xmlutil.NewEncoder (same as buildDocumentXML) ...
}
```

---

### `wordingo/doc.go` (utility, documentation — NEW)

No code analog needed. Package doc comment. Pattern: follow `wordingo.go:1-11` package comment style (godoc overview with example code).

---

### `wordingo/roundtrip_test.go` (test, CRUD — NEW)

**Analog:** `internal/opc/opc_test.go` lines 97-145 (`DiffParts`), 364-384 (`TestRoundTrip`), 517-536 (`TestRoundTripRealFixtures`)

**Imports pattern** (opc_test.go:3-13):
```go
import (
    "archive/zip"
    "bytes"
    "errors"
    "io"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)
```

For `wordingo/roundtrip_test.go`:
```go
package wordingo

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/fabiomarini/wordingo/internal/opc"
)
```

**Core DiffParts pattern** (opc_test.go:101-145):
```go
func DiffParts(t *testing.T, a, b []byte) {
    t.Helper()
    manifest := func(name string) bool {
        return name == "[Content_Types].xml" || isRelsPath(name)
    }
    partsOf := func(data []byte) map[string][]byte {
        zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
        if err != nil { t.Fatal(err) }
        m := make(map[string][]byte, len(zr.File))
        for _, f := range zr.File {
            rc, _ := f.Open()
            payload, _ := io.ReadAll(rc)
            rc.Close()
            m[f.Name] = payload
        }
        return m
    }
    pa, pb := partsOf(a), partsOf(b)
    for name, payload := range pa {
        if manifest(name) { continue }
        other, ok := pb[name]
        if !ok { t.Errorf("part %s missing from saved package", name); continue }
        if !bytes.Equal(payload, other) {
            t.Errorf("part %s payload differs (%d vs %d bytes)", name, len(payload), len(other))
        }
    }
    for name := range pb {
        if _, ok := pa[name]; !ok {
            t.Errorf("part %s unexpected in saved package", name)
        }
    }
}
```

**Phase 3 round-trip test** (from RESEARCH.md lines 381-403):
```go
func TestRoundTrip_MultiHeading(t *testing.T) {
    data, err := os.ReadFile("testdata/roundtrip/multi-heading.docx")
    if err != nil { t.Fatal(err) }

    doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
    if err != nil { t.Fatal(err) }

    paragraphs := doc.Paragraphs()
    if len(paragraphs) == 0 {
        t.Error("expected paragraphs in fixture")
    }

    var buf bytes.Buffer
    if _, err := doc.WriteTo(&buf); err != nil {
        t.Fatal(err)
    }

    opc.DiffParts(t, data, buf.Bytes())
}
```

**Core lazy-loading test** (from RESEARCH.md lines 408-431):
```go
func TestLazyLoading(t *testing.T) {
    data, err := os.ReadFile("testdata/roundtrip/header-footer.docx")
    if err != nil { t.Fatal(err) }

    doc, err := OpenReader(bytes.NewReader(data), int64(len(data)))
    if err != nil { t.Fatal(err) }

    if doc.doc == nil || doc.doc.Body == nil {
        t.Fatal("body should be eagerly parsed")
    }

    headerPart, ok := doc.pkg.Parts["word/header1.xml"]
    if !ok { t.Skip("no header part in fixture") }
    if headerPart.Modified() {
        t.Error("header part should not be marked modified")
    }
}
```

---

### `wordingo/template_test.go` (test, CRUD — NEW)

**Analog:** `internal/style/cloner_test.go` lines 242-264 (`TestCloneStyles_FreshEmptyTarget`), 395-420 (`TestCloneStyles_SaveRoundTrip`)

**Imports pattern** (cloner_test.go:3-14):
```go
import (
    "archive/zip"
    "bytes"
    "errors"
    "io"
    "os"
    "path/filepath"
    "strings"
    "testing"
    "time"
)
```

**Core pattern** — build synthetic source + target, call clone/FromTemplate, verify:
```go
func TestFromTemplate_Fresh(t *testing.T) {
    // Create a template document
    templatePath := "testdata/roundtrip/multi-heading.docx"
    doc, err := FromTemplate(templatePath)
    if err != nil { t.Fatal(err) }

    // Verify body is empty (CREATE-03)
    if len(doc.Paragraphs()) != 0 {
        t.Error("FromTemplate should clear body")
    }

    // Save and verify style parts survive
    var buf bytes.Buffer
    if _, err := doc.WriteTo(&buf); err != nil {
        t.Fatal(err)
    }

    // Re-open and check style parts present
    saved, err := OpenReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
    if err != nil { t.Fatal(err) }
    // ... verify parts ...
}
```

---

### `wordingo/create_test.go` (test, CRUD — MODIFY)

**Self-analog:** Existing `create_test.go` lines 1-172

**Core changes:**
- `doc.Save(&buf)` → `doc.WriteTo(&buf)` (line 17)
- `doc.Save(&buf)` → `doc.WriteTo(&buf)` (line 184 — second call)
- Add tests for `Save(path)` new convenience method

---

### `testdata/roundtrip/` (fixture, testdata — NEW)

**Analog:** `testdata/word/` directory with existing `.docx` fixtures

**Pattern:** Place 3-4 minimal .docx files per D-08:
- `testdata/roundtrip/blank.docx`
- `testdata/roundtrip/single-paragraph.docx`
- `testdata/roundtrip/multi-heading.docx`
- `testdata/roundtrip/header-footer.docx`

Fixtures need to be committed as binary. Tests reference by relative path from project root.

---

## Shared Patterns

### Error Wrapping (ALL Go files)
**Source:** `internal/opc/errors.go` patterns + `internal/style/errors.go`
```go
// Every error wraps a sentinel via fmt.Errorf("...: %w")
var ErrInvalidPackage = errors.New("opc: invalid package")
// Usage:
return nil, fmt.Errorf("wordingo: decode document.xml: %w", err)
```
**Apply to:** All new `.go` files — use `fmt.Errorf("wordingo: %s: %w")` pattern with meaningful context.

### X() Escape Hatch (all public API types)
**Source:** `wordingo.go:53-57`
```go
func (d *Document) X() *opc.Package {
    return d.pkg
}
```
**Apply to:** `paragraph.go` — `Paragraph.X() *wml.CT_P`

### Wrapper-with-Nil-Guard Accessor (all read-only accessors)
**Source:** `wordingo.go` and RESEARCH.md Pattern 1
```go
func (p *Paragraph) Style() string {
    if p.ct.PPr != nil && p.ct.PPr.PStyle != nil && p.ct.PPr.PStyle.Val != nil {
        return *p.ct.PPr.PStyle.Val
    }
    return ""
}
```
**Apply to:** All accessor methods on `Paragraph` — nil-check each pointer chain level.

### Deterministic ZIP Test Fixture Builder
**Source:** `internal/style/cloner_test.go:52-65` (`addZipEntry`)
```go
func addZipEntry(t *testing.T, zw *zip.Writer, name string, payload []byte) {
    t.Helper()
    h := &zip.FileHeader{Name: name, Method: zip.Deflate}
    h.Modified = time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC)
    w, err := zw.CreateHeader(h)
    if err != nil { t.Fatal(err) }
    if _, err := w.Write(payload); err != nil { t.Fatal(err) }
}
```
**Apply to:** `template_test.go` and `roundtrip_test.go` if synthetic fixture tests needed.

### Centralized Sentinel Errors
**Source:** `internal/opc/errors.go`, `internal/style/errors.go`
**Apply to:** If Document/Paragraph need new sentinel errors, place in a `wordingo/errors.go` file following the same pattern.

---

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `wordingo/doc.go` | utility | documentation | Pure godoc — no code analog needed |
| `testdata/roundtrip/` | fixture | testdata | Binary fixture dir — no Go code analog |

---

## Metadata

**Analog search scope:** `wordingo/`, `internal/opc/`, `internal/style/`, `internal/wml/`, `internal/xmlutil/`, `testdata/`
**Files scanned:** 18 Go source files + testdata dirs
**Pattern extraction date:** 2026-07-25
