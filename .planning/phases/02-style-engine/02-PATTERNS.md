# Phase 2: Style Engine - Pattern Map

**Mapped:** 2026-07-25
**Files analyzed:** 14 (7 new source, 5 new test, 1 modification, 6 fixture groups)
**Analogs found:** 12 / 14 (2 fixtures have no code analog — user-authored per Phase 1 D-02)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/style/errors.go` | config (sentinel) | n/a | `internal/opc/errors.go` | exact |
| `internal/style/cloner.go` | service | file-I/O (byte copy + rels/ct wiring) | `internal/opc/package.go` (`MarkModified`, `Save` loop) + `internal/opc/relationships.go` (`NextRID`) + `internal/opc/contenttypes.go` (`serialize`) | exact (composition of 3 Phase 1 APIs) |
| `internal/style/resolver.go` | service | transform (chain walk + memo) | no direct analog — greenfield; closest primitives: `internal/wml/styles.go` (merge target types) + `internal/opc/package.go:108` (`Warnings()` pattern) | role-match (new logic over Phase 1 types) |
| `internal/style/theme.go` | service | transform (parse-on-demand + color math) | `internal/xmlutil/decoder.go` (`SafeDecoder`) + `internal/wml/document.go:12` (`CT_Theme` RawXML hoard) | role-match (token-scan over RawXML-hoarded part) |
| `internal/style/numbering.go` | service | transform (numId+ilvl lookup + lvlOverride) | `internal/wml/numbering.go` (types) + `internal/xmlutil/decoder.go` (lazy parse) | role-match |
| `internal/wml/numbering.go` | **modification** | n/a (struct field add) | itself — extends `CT_Num` (lines 26-31) with `LvlOverride` per RESEARCH Pitfall 4 | exact (in-place extension) |
| `internal/style/cloner_test.go` | test | request-response | `internal/opc/opc_test.go` (`TestSave`, `TestRoundTrip`, `DiffParts`) | exact |
| `internal/style/resolver_test.go` | test | transform | `internal/opc/opc_test.go` (`TestOpen`, `TestSafety` subtest pattern) + `internal/wml/wml_test.go` (compile-check + round-trip) | role-match |
| `internal/style/theme_test.go` | test | transform | `internal/wml/wml_test.go` (marshal/unmarshal round-trip) | role-match |
| `internal/style/numbering_test.go` | test | transform | `internal/wml/wml_test.go` | role-match |
| `internal/style/testdata_test.go` | test (harness) | file-I/O | `internal/opc/opc_test.go` (`HasFixtures`, `realFixtures`, `buildSyntheticZip`) | exact |
| `testdata/word/style-rich/*.docx` + `*.expected.json` | fixture | file-I/O | `testdata/word/*.docx` (Phase 1 D-02 corpus) | partial (docx analog exists; expected.json schema is new) |
| `testdata/style-engine/hostile/*.docx` + `*.expected.json` | fixture | file-I/O | `testdata/hostile/` (Phase 1 D-02 hostile corpus) | partial |
| `internal/opc/package.go` (possible `Part.IsModified()` accessor) | **modification** | n/a | itself line 45 (`modified` field) | exact (one-liner accessor if needed) |

## Pattern Assignments

### `internal/style/errors.go` (sentinel error)

**Analog:** `internal/opc/errors.go` (lines 1-27)

**Pattern — sentinel + `%w` wrapping (Phase 1 D-11):**
```go
// Source: internal/opc/errors.go:1-27
package opc

import "errors"

var (
    ErrInvalidPackage    = errors.New("opc: invalid package")
    ErrUnsafePath        = errors.New("opc: unsafe path")
    ErrDecompressionLimit = errors.New("opc: decompression limit exceeded")
    ErrTooManyParts      = errors.New("opc: too many parts")
    ErrXMLDepth          = errors.New("opc: xml depth limit exceeded")
)
```
**Copy for `internal/style/errors.go`:**
```go
package style

import "errors"

// ErrCloneTargetNotEmpty marks a clone target that already has one of
// the 5 style parts (D-08). Callers match with errors.Is.
var ErrCloneTargetNotEmpty = errors.New("style: clone target not empty")
```
**Wrapping convention (from `internal/opc/package.go:53`, `:60`, `:79`):** every returned error wraps a sentinel via `fmt.Errorf("style: <context>: %w", err)`. No typed error structs (Phase 1 D-11).

---

### `internal/style/cloner.go` (service, file-I/O byte copy)

**Analogs (compose 3 Phase 1 APIs):**
- `internal/opc/package.go` — `MarkModified`, `Parts` map, `Part.Open`
- `internal/opc/relationships.go` — `NextRID`, `Relationship` struct
- `internal/opc/contenttypes.go` — `ContentTypes.Overrides`

**Imports pattern** (mirror `internal/opc/package.go:10-17`):
```go
package style

import (
    "fmt"
    "github.com/fabiomarini/wordingo/internal/opc"
)
```

**Core pattern — `MarkModified` byte-copy** (`internal/opc/package.go:113-121`):
```go
// Source: internal/opc/package.go:113-121
func (p *Package) MarkModified(name string, data []byte) {
    part, ok := p.Parts[name]
    if !ok {
        part = &Part{Name: name}
        p.Parts[name] = part
    }
    part.data = data
    part.modified = true
}
```
Cloner calls `dst.MarkModified(partName, srcBytes)` per source part — this is the entire byte pass-through path (D-09). No re-parse.

**Read source bytes** (`internal/opc/package.go:219-230`):
```go
// Source: internal/opc/package.go:219-230 — readAllCapped
func readAllCapped(p *Part) ([]byte, error) {
    rc, err := p.Open()
    if err != nil { return nil, err }
    defer rc.Close()
    b, err := io.ReadAll(rc)
    if err != nil {
        return nil, fmt.Errorf("opc: read part %s: %w", p.Name, err)
    }
    return b, nil
}
```
`readAllCapped` is unexported in `opc`. Cloner either (a) calls `src.Parts[name].Open()` + `io.ReadAll` directly, or (b) RESEARCH §Cache Invalidation Open Question 3 recommends a one-liner accessor. **Recommended: cloner uses `src.Parts[name].Open()` + `io.ReadAll` inline (~5 LOC) to avoid touching `opc` export surface.**

**rId allocation** (`internal/opc/relationships.go:138-145`):
```go
// Source: internal/opc/relationships.go:138-145
func (rs *Relationships) NextRID() string {
    if rs.next == 0 { rs.rescan() }
    id := "rId" + strconv.Itoa(rs.next)
    rs.next++
    return id
}
```
Cloner: `wordRels := dst.Rels["word/document.xml"]; if wordRels == nil { wordRels = &opc.Relationships{}; dst.Rels["word/document.xml"] = wordRels }; rid := wordRels.NextRID(); wordRels.Rels = append(wordRels.Rels, opc.Relationship{ID: rid, Type: relType, Target: partName})`.

**Content-type override** (`internal/opc/contenttypes.go:17-20, 60-75`):
```go
// Source: internal/opc/contenttypes.go:17-20
type ContentTypes struct {
    Defaults  map[string]string // extension (no dot) -> MIME
    Overrides map[string]string // part name (leading /) -> MIME
}
```
`ContentTypes.Overrides` is exported — cloner sets `dst.ContentTypes.Overrides["/"+partName] = ct` directly. `TypeFor` (`:60-75`) reads `["/"+partName]` first. **Note: key has leading `/` — RESEARCH Pattern 5 uses no leading slash in `cloneParts`; reconcile by writing `dst.ContentTypes.Overrides["/"+p.name] = p.ct`.**

**Target-empty precondition** (D-08): check `if _, exists := dst.Parts[p.name]; exists { return ErrCloneTargetNotEmpty }` BEFORE `MarkModified`. Mirror `internal/opc/package.go:175-178` (mandatory-part guard) for the shape:
```go
// Source: internal/opc/package.go:175-178 — mandatory-part guard
ctPart, ok := pkg.Parts["[Content_Types].xml"]
if !ok {
    return nil, fmt.Errorf("opc: missing [Content_Types].xml: %w", ErrInvalidPackage)
}
```

**Error wrapping** (`internal/opc/package.go:53`):
```go
return nil, fmt.Errorf("opc: part %s deleted: %w", p.Name, ErrInvalidPackage)
```
Cloner: `return fmt.Errorf("style: clone %s: %w", p.name, ErrCloneTargetNotEmpty)`.

**Part-name + content-type table** (from RESEARCH Pattern 5, lines 349-359 — concrete constants to hardcode in `cloner.go`):
```go
var cloneParts = []struct{ name, relType, ct string }{
    {"word/styles.xml",         "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles",    "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"},
    {"word/numbering.xml",      "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering", "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"},
    {"word/fontTable.xml",      "http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable",  "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"},
    {"word/theme/theme1.xml",   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme",      "application/vnd.openxmlformats-officedocument.theme+xml"},
    {"word/settings.xml",       "http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings",  "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"},
}
```

**Path-traversal safety (inherited)** (`internal/opc/package.go:352-380`): `validatePartName` already runs on `Open`. Cloner uses fixed constants above (no user-supplied names) — no new validation needed (RESEARCH Security Domain V12).

---

### `internal/style/resolver.go` (service, transform — chain walk + memo)

**Analog for return type + escape hatch:** `internal/wml/document.go:60-75` (`CT_PPr`) + `:89-110` (`CT_RPr`). Resolver returns `*wml.CT_PPr` / `*wml.CT_RPr` clones (D-01 — no parallel `EffectiveProps`).

**Imports pattern:**
```go
package style

import (
    "encoding/xml"
    "fmt"
    "github.com/fabiomarini/wordingo/internal/opc"
    "github.com/fabiomarini/wordingo/internal/wml"
    "github.com/fabiomarini/wordingo/internal/xmlutil"
)
```

**Merge target types** (`internal/wml/styles.go:36-106`):
```go
// Source: internal/wml/styles.go:37-43 — CT_Styles (resolver parses this lazily)
type CT_Styles struct {
    XMLName      xml.Name         `xml:"... styles"`
    DocDefaults  *CT_DocDefaults  `xml:"... docDefaults"`
    LatentStyles *CT_LatentStyles `xml:"... latentStyles"`
    Style        []*CT_Style      `xml:"... style"`
    Raw          []xmlutil.RawXML `xml:",any"`
}
// Source: internal/wml/styles.go:46-60 — CT_Style (chain node)
type CT_Style struct {
    StyleID  *string  `xml:"... styleId,attr,omitempty"`
    BasedOn  *CT_BasedOn `xml:"... basedOn"`   // walk this (Pattern 6)
    Next     *CT_Next    `xml:"... next"`      // NEVER walk (Pitfall 1)
    Link     *CT_Link    `xml:"... link"`      // optional rPr layer (Open Q 1)
    PPr      *CT_PPr   `xml:"... pPr"`
    RPr      *CT_RPr   `xml:"... rPr"`
}
// Source: internal/wml/styles.go:87-106 — DocDefaults layers
type CT_DocDefaults struct {
    RPrDefault *CT_RPrDefault // → .RPr
    PPrDefault *CT_PPrDefault // → .PPr
}
```

**Pointer-field nil-merge semantics (Pattern 1, Phase 1 D-07):** all fields on `CT_PPr`/`CT_RPr` are pointers (`internal/wml/document.go:60-110`). Merge rule:
- **Shallow override (default):** `if child.PPr.Spacing != nil { result.PPr.Spacing = clone(child.PPr.Spacing) }` — applies to all pointer fields except the 4 composite leaves.
- **Deep-merge (composite leaves):** `CT_Spacing` (`properties.go:24-32`), `CT_Ind` (`:35-41`), `CT_RFonts` (`:109-117`), `CT_Color` (`:134-140`) — per-attribute independence per ISO §17.7.2. RESEARCH §Pattern 1 (lines 218-240) gives the `mergeSpacing` template.

**Cycle detection (D-05) — warning surfacing** (`internal/opc/package.go:103-108`):
```go
// Source: internal/opc/package.go:103-108 — Warnings() pattern (Phase 1 D-12)
type Package struct {
    warnings []string
}
func (p *Package) Warnings() []string { return p.warnings }
```
Resolver holds its own `warnings []string` and appends via `warn(fmt.Sprintf("style: circular basedOn chain at %q", cursor))` (RESEARCH Pattern 2, lines 250-258). Phase 4 `Document` merges opc + style warnings.

**Memo cache (D-03) + dirty-flag invalidation (RESEARCH §Cache Invalidation):** read `opc.Part.modified` (`internal/opc/package.go:45`) — RESEARCH Open Q 3 / Assumption A5: if unexported, add `Part.IsModified() bool` one-liner to `internal/opc/package.go`. Resolver's `ResolvePPr` entry:
```go
// Pseudocode anchored to internal/opc/package.go:45 (Part.modified field)
if r.pkg.Parts["word/styles.xml"].modified {  // or .IsModified()
    r.memoPPr = nil
    r.memoRPr = nil
}
```
**Flag for planner:** verify `Part.modified` export status in plan 02-01; if unexported, add accessor (1-line `opc` modification).

**Clone-before-return (D-01/D-03, Pitfall 8):** every return path does `clonePPr(cached)` — never return `memo[styleId]` directly. Deep-clone via field-by-field copy of pointer fields (reuse the same `mergePPr`/`mergeSpacing` helpers used for chain merge — a clone is a merge into an empty `&CT_PPr{}`).

**Chain walk** (RESEARCH §Code Examples, lines 480-529 — canonical `resolvePPr` + `collectChain`): walk `style.BasedOn.Val` (`internal/wml/styles.go:69-72`, `Val *string`); consult `latentStyles` only when `findStyle(cursor) == nil` (D-12); `next` is NEVER read (Pitfall 1).

---

### `internal/style/theme.go` (service, transform — parse-on-demand color map)

**Analog for lazy parse over RawXML-hoarded part:** `internal/wml/document.go:10-15` (`CT_Theme` with `Raw []xmlutil.RawXML`) + `internal/xmlutil/decoder.go` (`SafeDecoder`).

**RawXML hoarding pattern** (`internal/xmlutil/rawxml.go:19-44`):
```go
// Source: internal/xmlutil/rawxml.go:19-44 — RawXML captures subtree token-by-token
type RawXML struct { Tokens []xml.Token }
func (r *RawXML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
    r.Tokens = append(r.Tokens[:0], xml.CopyToken(start))
    depth := 1
    for depth > 0 {
        tok, err := d.Token()
        if err != nil { return fmt.Errorf("rawxml: %w", err) }
        tok = xml.CopyToken(tok)
        switch tok.(type) {
        case xml.StartElement: depth++
        case xml.EndElement: depth--
        }
        r.Tokens = append(r.Tokens, tok)
    }
    return nil
}
```
Theme.go does NOT use `CT_Theme` (full DrawingML out of scope, RESEARCH anti-pattern). Instead, token-scan `theme1.xml` bytes directly via `xmlutil.NewSafeDecoder` (`internal/xmlutil/decoder.go:36-41`):
```go
// Source: internal/xmlutil/decoder.go:36-41
func NewSafeDecoder(r io.Reader, maxBytes int64) *SafeDecoder {
    d := xml.NewDecoder(io.LimitReader(r, maxBytes))
    d.Strict = true
    d.Entity = nil
    return &SafeDecoder{Decoder: d}
}
```
Scan for `<a:clrScheme>` children (`dk1/lt1/dk2/lt2/accent1..6/hlink/folHlink`), read `<a:srgbClr val="..."/>` or `<a:sysClr ... lastClr="..."/>` payloads. Verified element names against `defaults/theme1.xml` line 2: `<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>` etc.

**Enum → element-name map (RESEARCH Pattern 3, lines 269-282 — the critical mismatch):**
```go
var themeEnumToElement = map[string]string{
    "dark1": "dk1", "light1": "lt1", "dark2": "dk2", "light2": "lt2",
    "accent1": "accent1", "accent2": "accent2", "accent3": "accent3",
    "accent4": "accent4", "accent5": "accent5", "accent6": "accent6",
    "hyperlink": "hlink", "followedHyperlink": "folHlink",
}
```

**Color struct to mutate** (`internal/wml/properties.go:134-140`):
```go
// Source: internal/wml/properties.go:134-140
type CT_Color struct {
    Val        *string `xml:"... val,attr,omitempty"`
    ThemeColor *string `xml:"... themeColor,attr,omitempty"`
    ThemeShade *string `xml:"... themeShade,attr,omitempty"`
    ThemeTint  *string `xml:"... themeTint,attr,omitempty"`
}
```
`ResolveColor` (RESEARCH lines 556-578): shallow-copy input, if `ThemeColor != nil` → lookup element name → lookup hex → apply `ThemeShade` then `ThemeTint` (Pitfall 5: shade-first) → set `Val`, nil-out `ThemeColor/ThemeShade/ThemeTint`. On any miss → `warn(...)`, return copy as-is (D-07).

**Tint/shade math (RESEARCH Pattern 3, lines 284-294 — CITED, verify against fixture):**
- `tint = parsed/255; new = orig + (255-orig)*tint` (lighten)
- `shade = parsed/255; new = orig*shade` (darken)
- Both present: shade first, then tint.

**Warning surfacing:** same `warn func(string)` callback as resolver — append to resolver's `warnings []string`.

---

### `internal/style/numbering.go` (service, transform — numId+ilvl lookup)

**Analog:** `internal/wml/numbering.go` (types) + `internal/xmlutil/decoder.go` (lazy parse).

**Types to parse lazily** (`internal/wml/numbering.go:9-67`):
```go
// Source: internal/wml/numbering.go:9-15 — root
type CT_Numbering struct {
    Num         []*CT_Num         `xml:"... num"`
    AbstractNum []*CT_AbstractNum `xml:"... abstractNum"`
}
// Source: internal/wml/numbering.go:18-23 — abstract def
type CT_AbstractNum struct {
    AbstractNumID *int64    `xml:"... abstractNumId,attr,omitempty"`
    Lvl           []*CT_Lvl `xml:"... lvl"`
}
// Source: internal/wml/numbering.go:26-31 — instance (⚠ MISSING lvlOverride)
type CT_Num struct {
    NumID         *int64            `xml:"... numId,attr,omitempty"`
    AbstractNumID *CT_AbstractNumID `xml:"... abstractNumId"`
    Raw           []xmlutil.RawXML  `xml:",any"`
}
// Source: internal/wml/numbering.go:40-49 — level (the merge target)
type CT_Lvl struct {
    ILvl    *int64       `xml:"... ilvl,attr,omitempty"`
    NumFmt  *CT_NumFmt   `xml:"... numFmt"`
    LvlText *CT_LvlText  `xml:"... lvlText"`
    Start   *CT_Start    `xml:"... start"`
    PPr     *CT_PPr      `xml:"... pPr"`   // merged into effective pPr (D-13)
    RPr     *CT_RPr      `xml:"... rPr"`
}
```

**Resolve algorithm (RESEARCH Pattern 4, lines 324-339):**
1. `findNum(numId)` → `CT_Num` (match `NumID`)
2. read `num.AbstractNumID.Val` (`internal/wml/numbering.go:34-37`)
3. `findAbstractNum(abstractNumId)` → `CT_AbstractNum` (match `AbstractNumID`)
4. `findLvl(abs, ilvl)` → `CT_Lvl` (match `ILvl`)
5. Apply `lvlOverride` if present (requires the struct fix below)

**Merge level pPr into effective pPr (D-13):** `mergePPr(result, lvl.PPr)` using the same nil-merge helpers as resolver (Pattern 1). Indentation fields (`CT_Ind.Left/Right/FirstLine/Hanging`, `properties.go:35-41`) are composite-leaf deep-merge.

---

### `internal/wml/numbering.go` (**modification** — extend `CT_Num`)

**Analog:** itself — in-place field addition. RESEARCH Pitfall 4 (lines 446-449) + Assumption A4.

**Current struct** (`internal/wml/numbering.go:26-31`):
```go
type CT_Num struct {
    XMLName        xml.Name        `xml:"... num"`
    NumID          *int64          `xml:"... numId,attr,omitempty"`
    AbstractNumID  *CT_AbstractNumID `xml:"... abstractNumId"`
    Raw            []xmlutil.RawXML `xml:",any"`
}
```
**Required extension (RESEARCH lines 315-323):**
```go
type CT_Num struct {
    XMLName       xml.Name          `xml:"... num"`
    NumID         *int64            `xml:"... numId,attr,omitempty"`
    AbstractNumID *CT_AbstractNumID `xml:"... abstractNumId"`
    LvlOverride   []*CT_LvlOverride `xml:"... lvlOverride"`  // NEW — ISO §17.9.18
    Raw           []xmlutil.RawXML  `xml:",any"`
}

// NEW types (ISO §17.9.18 lvlOverride, §17.9.24 startOverride)
type CT_LvlOverride struct {
    XMLName      xml.Name        `xml:"... lvlOverride"`
    ILvl         *int64          `xml:"... ilvl,attr,omitempty"`
    StartOverride *CT_StartOverride `xml:"... startOverride"`
    Lvl          *CT_Lvl         `xml:"... lvl,omitempty"`
}

type CT_StartOverride struct {
    XMLName xml.Name `xml:"... startOverride"`
    Val     *int64   `xml:"... val,attr,omitempty"`
}
```
**Backward compat (Assumption A4):** fixtures without `<w:lvlOverride>` parse identically — the new typed field just captures what `Raw` previously hoarded. **No fixture regeneration** (RESEARCH Runtime State Inventory: `defaults/` has no numbering.xml; verified).

**Pattern to copy for struct tags:** every field uses the full URI tag form `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main <local>"` (URI-based, not prefix — Phase 1 OPC-03, matches every struct in `internal/wml/*.go`).

---

### `internal/style/*_test.go` (tests)

**Analog:** `internal/opc/opc_test.go` (585 lines — the canonical test file).

**Synthetic fixture builder** (`internal/opc/opc_test.go:47-73`):
```go
// Source: internal/opc/opc_test.go:47-73 — buildSyntheticZip
func buildSyntheticZip(t *testing.T) []byte {
    t.Helper()
    var buf bytes.Buffer
    zw := zip.NewWriter(&buf)
    add := func(name string, payload []byte) {
        t.Helper()
        h := &zip.FileHeader{Name: name, Method: zip.Deflate}
        h.Modified = time.Date(2020, 3, 4, 5, 6, 7, 0, time.UTC)
        w, err := zw.CreateHeader(h)
        if err != nil { t.Fatal(err) }
        if _, err := w.Write(payload); err != nil { t.Fatal(err) }
    }
    add("[Content_Types].xml", []byte(synContentTypes))
    // ...
    return buf.Bytes()
}
```
Cloner test: build a source zip with the 5 style parts + a fresh-empty target zip, call `CloneStyles`, `Save`, unzip both, byte-diff the 5 parts. **Reuse `DiffParts`** (`internal/opc/opc_test.go:101-145`) — but it's in package `opc`. Either (a) copy the helper into `style_test.go`, or (b) export it. **Recommendation: copy the helper** (keeps `opc` export surface unchanged).

**`errors.Is` assertion** (`internal/opc/opc_test.go:240-242`):
```go
// Source: internal/opc/opc_test.go:240-242
if !errors.Is(err, ErrUnsafePath) {
    t.Fatalf("err = %v, want ErrUnsafePath", err)
}
```
Cloner conflict test: `if !errors.Is(err, ErrCloneTargetNotEmpty) { t.Fatalf(...) }`.

**Fixture-presence skip** (`internal/opc/opc_test.go:86-95`):
```go
// Source: internal/opc/opc_test.go:86-95 — HasFixtures skip pattern
func HasFixtures(t *testing.T) bool {
    t.Helper()
    _, err := os.Stat(filepath.Join("..", "..", "testdata", "word", "blank.docx"))
    if err != nil {
        t.Skip("testdata/ fixture corpus not present (D-02 pending) — synthetic path active")
        return false
    }
    return true
}
```
`testdata_test.go` (resolver harness) follows this exact pattern: check `testdata/word/style-rich/*.expected.json` presence, skip if absent (user-authored per Phase 1 D-02). Synthetic tests (over `defaults/*.xml`) run unconditionally.

**Subtest structure** (`internal/opc/opc_test.go:212-287` — `TestSafety` with `t.Run` per case): use for hostile-fixture table (circular-basedon, dangling-basedon, missing-numid).

**No-panic guard** (`internal/opc/opc_test.go:573-583`):
```go
// Source: internal/opc/opc_test.go:573-583
defer func() {
    if r := recover(); r != nil {
        t.Fatalf("Open panicked on %s: %v", c.name, r)
    }
}()
```
Apply to resolver tests against hostile fixtures (circular basedOn must not stack-overflow — RESEARCH Security Domain DoS mitigation).

---

### `testdata/word/style-rich/` + `testdata/style-engine/hostile/` (fixtures)

**Analog:** `testdata/word/*.docx` (Phase 1 D-02 — user-authored, 6 files present: `blank.docx`, `01-blank.docx`..`05-...`) + `testdata/hostile/` (Phase 1 hostile corpus).

**Expected-values JSON schema (RESEARCH §Fixture Design, lines 599-639):** new schema, no codebase analog. Planner: this is a user-authored artifact per Phase 1 D-02 pattern — implementation tasks build the harness + synthetic fixtures; real `.docx` files are authored in Word (not on this machine, RESEARCH §Environment Availability). `expected.json` fields: `null` = absent (nil pointer) ≠ `0`/`""` = zero value; `warnings` array asserts exact substrings.

**Hostile dir layout (RESEARCH lines 657-679):** sibling `testdata/style-engine/hostile/` keeps golden `style-rich/` equality assertions clean from warning-string noise.

---

### `internal/opc/package.go` possible `Part.IsModified()` accessor (conditional modification)

**Analog:** itself line 45 (`modified bool` field, unexported).

**Trigger:** RESEARCH Open Q 3 / Assumption A5 — resolver cache invalidation needs to read `Part.modified`. If field is unexported (it is — `internal/opc/package.go:45` shows `modified bool` lowercase), add:
```go
// Add to internal/opc/package.go near line 108 (Warnings accessor)
func (p *Part) IsModified() bool { return p.modified }
```
**Planner note:** this is a 1-line `opc` modification gated on the resolver actually needing it. If resolver holds a reference to `*opc.Package` and the cache check is inside `opc`-aware code, an exported accessor is cleanest. Flag as Wave 0 prerequisite for plan 02-01.

## Shared Patterns

### Sentinel errors + `%w` wrapping (Phase 1 D-11)
**Source:** `internal/opc/errors.go:1-27` + `internal/opc/package.go:53,60,79,137,157,162,200,227,280,325,329,341,344,354`
**Apply to:** `internal/style/errors.go` (`ErrCloneTargetNotEmpty`), `internal/style/cloner.go` (wrap `ErrCloneTargetNotEmpty` + any `opc` errors via `%w`), `internal/style/resolver.go`/`theme.go`/`numbering.go` (wrap parse errors).
```go
// Convention: fmt.Errorf("<pkg>: <context>: %w", sentinel)
return fmt.Errorf("style: clone %s: %w", p.name, ErrCloneTargetNotEmpty)
```
No typed error structs. Callers use `errors.Is`.

### `Warnings() []string` (Phase 1 D-12)
**Source:** `internal/opc/package.go:103-108`
**Apply to:** `internal/style/resolver.go` — resolver holds `warnings []string`, exposes `Warnings() []string`. All D-05 (cycle), D-07 (missing ref/theme/numId) paths append via a `warn func(string)` callback threaded through `theme.ResolveColor` and `numbering.ResolveLvl`. Phase 4 `Document.Warnings()` merges `opc.Warnings()` + `style.Warnings()`.
```go
// Source: internal/opc/package.go:108
func (p *Package) Warnings() []string { return p.warnings }
```

### Pointer-field nil-merge (Phase 1 D-07)
**Source:** every struct in `internal/wml/*.go` — all optional fields are pointers. `internal/wml/document.go:60-75` (`CT_PPr`), `:89-110` (`CT_RPr`), `internal/wml/properties.go:24-41` (`CT_Spacing`, `CT_Ind`), `:109-117` (`CT_RFonts`), `:134-140` (`CT_Color`).
**Apply to:** `internal/style/resolver.go` (`mergePPr`, `mergeRPr` + deep-merge helpers for the 4 composite leaves), `internal/style/numbering.go` (level pPr merge into effective pPr).
**Rule (RESEARCH Pattern 1):** shallow-override for pointer fields by default; deep-merge only for `CT_Spacing`, `CT_Ind`, `CT_RFonts`, `CT_Color` (per-attribute independence per ISO §17.7.2). Pitfall 7: applying shallow-merge to `CT_Spacing` drops parent's `after` when child sets only `before`.

### URI-based namespace tags (Phase 1 OPC-03)
**Source:** every struct tag in `internal/wml/*.go` uses full URI form `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main <local>"` — NOT prefix-based. `internal/xmlutil/ns.go:10-55` (`CanonicalPrefixes`) + `internal/xmlutil/decoder.go` (SafeDecoder) handle prefix-agnostic parse.
**Apply to:** `internal/wml/numbering.go` modification (new `CT_LvlOverride`/`CT_StartOverride` structs must use URI tags), any new structs in `internal/style/` (none expected — style reuses `wml` types per D-01).

### Lazy parse via `xmlutil.SafeDecoder` (Phase 1)
**Source:** `internal/xmlutil/decoder.go:36-65` (`NewSafeDecoder`, `Decode`), `internal/xmlutil/rawxml.go` (`RawXML` hoarding for unmodeled elements).
**Apply to:** `internal/style/theme.go` (token-scan `theme1.xml` for `a:clrScheme`), `internal/style/numbering.go` (decode `numbering.xml` → `wml.CT_Numbering`), `internal/style/resolver.go` (decode `styles.xml` → `wml.CT_Styles`). All parse on first query, cache tree, invalidate on `Part.modified`.

### OPC pass-through byte copy (Phase 1 OPC-04)
**Source:** `internal/opc/package.go:113-121` (`MarkModified`), `:317-326` (`Save` raw-copy branch: `zw.Copy(part.file)`).
**Apply to:** `internal/style/cloner.go` — the entire clone is `MarkModified` + `NextRID` + `ContentTypes.Overrides` wiring. No WML parse in cloner (D-09). Byte-identity inherited from `opc.Save` raw-copy of unmodified-original + deterministic-header write of modified parts.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `testdata/word/style-rich/*.expected.json` | fixture (schema) | file-I/O | New expected-values JSON schema (RESEARCH lines 599-639). No codebase analog — user-authored per Phase 1 D-02. Planner: build the harness loader in `testdata_test.go`; the `.docx` + `.json` pairs are authored artifacts (Word not on this machine per RESEARCH §Environment Availability). |
| `testdata/style-engine/hostile/*.expected.json` | fixture (schema) | file-I/O | Same — hostile expected-values schema is new. Synthetic hostile fixtures (built in-test via `zip.Writer`, analog `opc_test.go:47-73`) cover the warning-path assertions; real `.docx` hostile files are user-authored. |

## Metadata

**Analog search scope:**
- `internal/opc/` (6 files: package.go, relationships.go, contenttypes.go, errors.go, zipio.go, opc_test.go)
- `internal/wml/` (7 files: document.go, styles.go, numbering.go, properties.go, table.go, namespaces.go, wml_test.go)
- `internal/xmlutil/` (5 files: decoder.go, encoder.go, ns.go, rawxml.go, xmlutil_test.go)
- `defaults/` (5 XML files: styles.xml, theme1.xml, fontTable.xml, settings.xml, webSettings.xml)
- `testdata/` (word/, libreoffice/, googledocs/, hostile/)
- `.planning/phases/01-foundation/01-CONTEXT.md` (D-11/D-12 pattern anchors)

**Files scanned:** 23 source/test files + 5 default XMLs + planning docs
**Pattern extraction date:** 2026-07-25

**Key anchors for planner:**
- `internal/opc/package.go:45` — `Part.modified` field (export status check for resolver cache invalidation)
- `internal/opc/package.go:108` — `Warnings()` accessor shape to copy in resolver
- `internal/opc/package.go:113-121` — `MarkModified` (the cloner's core call)
- `internal/opc/relationships.go:138-145` — `NextRID` (cloner rId allocation)
- `internal/opc/contenttypes.go:17-20` — `ContentTypes.Overrides` map (leading `/` key)
- `internal/opc/errors.go:7-27` — sentinel pattern to copy for `ErrCloneTargetNotEmpty`
- `internal/opc/opc_test.go:47-73` — `buildSyntheticZip` fixture builder to copy into cloner_test
- `internal/opc/opc_test.go:101-145` — `DiffParts` byte-diff harness to copy into cloner_test
- `internal/opc/opc_test.go:240-242` — `errors.Is` assertion shape for conflict test
- `internal/opc/opc_test.go:573-583` — no-panic guard for hostile resolver tests
- `internal/wml/styles.go:37-106` — `CT_Styles`/`CT_Style`/`CT_DocDefaults` (resolver merge targets + chain walk)
- `internal/wml/numbering.go:26-31` — `CT_Num` struct to extend with `LvlOverride`
- `internal/wml/document.go:60-110` — `CT_PPr`/`CT_RPr` field lists (resolver return types + nil-merge targets)
- `internal/wml/properties.go:24-41, 109-117, 134-140` — `CT_Spacing`/`CT_Ind`/`CT_RFonts`/`CT_Color` (deep-merge composite leaves)
- `internal/xmlutil/decoder.go:36-41` — `NewSafeDecoder` for lazy theme/numbering/styles parse
- `internal/xmlutil/rawxml.go:19-44` — `RawXML` hoarding (theme.go avoids full `CT_Theme` parse)
- `defaults/theme1.xml:2` — verified `a:clrScheme` element names: `dk1/lt1/dk2/lt2/accent1-6/hlink/folHlink` with `srgbClr`/`sysClr` payloads
