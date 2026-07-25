# Phase 1: Foundation - Research

**Researched:** 2026-07-25
**Domain:** OPC package layer (ECMA-376 Part 2), Go `encoding/xml` namespace handling, WordprocessingML type modeling, blank .docx generation
**Confidence:** HIGH (core mechanics verified; safety-limit numbers are proposed values at discretion)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Verification = committed fixture corpus + golden files in CI; manual Word-open check gates phase sign-off. No Open XML SDK validator container.
- **D-02:** User authors real .docx fixtures (Word, LibreOffice, Google Docs) committed under `testdata/`.
- **D-03:** Fixture set 6–8 files: `blank.docx` + `styled.docx` per producer, plus one hostile file (customXml + glossary part) for OPC-04 pass-through tests.
- **D-04:** Round-trip fidelity asserted as per-part byte diff (unzip both, diff each part), not whole-file ZIP bytes.
- **D-05:** xmlutil = token-stream wrapper over `encoding/xml` Decoder/Encoder: normalize prefixes→URIs on read, emit canonical prefixes on write. WML structs keep `encoding/xml` struct tags. No fully custom parser.
- **D-06:** WML-04 unknown-child hoarding via raw token subtree blobs (`xmlutil.RawXML`-type field per struct), re-emitted verbatim. Not a generic element tree.
- **D-07:** WML struct fields use pointers for optional attributes/children (nil = absent). Value structs, no presence flags.
- **D-08:** Write canonical OOXML prefixes (`w`, `r`, `a`, `wp`, …); read accepts any prefix resolved by namespace URI.
- **D-09:** Module path `github.com/fabiomarini/wordingo`.
- **D-10:** Go version floor 1.23 in go.mod; CI tests 1.23 + latest stable.
- **D-11:** Error taxonomy = sentinel errors + `fmt.Errorf("...: %w", err)`; callers use `errors.Is`. No typed error structs.
- **D-12:** `Warnings()` returns `[]string`.

### the agent's Discretion
Internal package organization within `internal/opc`, `internal/wml`, `internal/xmlutil`; exact list of ~60 WML types; test naming/layout; safety-limit numeric thresholds (OPC-07) — proposed below.

### Deferred Ideas (OUT OF SCOPE)
- Open XML SDK validator in CI (.NET container)
- V2 features (field codes, comments, bookmarks, SDT, tracked changes, charts, equations)
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OPC-01 | Open .docx as OPC package ([Content_Types].xml, _rels/.rels, per-part .rels) | §Architecture Patterns (OPC reader), §Code Examples |
| OPC-02 | Canonical ZIP entry ordering on write (ECMA-376 Part 2 §9.1.4.2) | §Architecture Patterns → canonical ordering table |
| OPC-03 | Namespace resolution by URI, 40+ namespaces, producer-identical parsing | §xmlutil design, namespace registry table |
| OPC-04 | Pass-through preservation of unmodeled parts byte-identical | §Standard Stack (`zip.Writer.Copy` / `OpenRaw`), §Pitfall 5 |
| OPC-05 | Read Transitional + Strict; write Transitional; match source when editing | §Conformance section |
| OPC-06 | Relationship graph integrity (unique rId, no reuse, validation on save) | §Relationship graph design |
| OPC-07 | Safety limits (zip bomb, entity expansion, path traversal, max parts) | §Safety Limits (proposed numbers) |
| WML-01 | ~60 essential WML struct types | §WML Type List (62 types proposed) |
| WML-02 | Bidirectional marshal/unmarshal, correct namespace URIs | §xmlutil design, §Code Examples |
| WML-03 | Whitespace fidelity (xml:space="preserve"; no implicit run merging) | §Pitfall: whitespace, §Code Examples |
| WML-04 | Unknown children hoarded and re-emitted | §RawXML design (D-06), §Code Examples |
| CREATE-01 | Blank .docx: Normal + Heading 1–9 + Title styles, default theme/fontTable/settings, one sectPr | §Blank Document Part Set |
| CREATE-02 | Blank passes Word validation without repair | §Blank Document Part Set + verification strategy |
</phase_requirements>

## Summary

Phase 1 decomposes into three plans matching the roadmap: (01-01) OPC package layer on `archive/zip`, (01-02) xmlutil namespace registry + WML types on `encoding/xml`, (01-03) blank-document generator assembling static default parts. The entire stack is Go stdlib (`archive/zip`, `encoding/xml`, `errors`, `fmt`, `io`) — zero external dependencies, so no package-legitimacy audit is required.

Two findings materially shape the plan. First, `archive/zip` in Go 1.17+ provides `File.OpenRaw` and `Writer.Copy`/`Writer.CreateRaw`, which copy compressed data without decompression/recompression — this makes OPC-04 pass-through byte-exactness cheap and also gives a clean decompression-bomb choke point (we control all decompression). Second, `encoding/xml` **does** resolve namespace prefixes to URIs on unmarshal when struct tags use the `xml:"URI local"` form (so OPC-03 read-side works out of the box), but on **marshal** it emits `xmlns="URI"` attributes rather than prefixed names — so the write side of xmlutil is the real work: a prefix-emitting encoder shim that rewrites `xml.Name{Space: URI}` into `prefix:local` with a single `xmlns:prefix` declaration set on the root. This fits the ~500 LOC budget; the token-stream design below estimates ~450–650 LOC total and should be flagged if it exceeds ~800.

**Primary recommendation:** Build xmlutil as two narrow layers — a read normalizer (trivial: encoding/xml already URI-resolves) and a write canonicalizer (token-to-token prefix emitter with an URI→prefix registry) — plus a `RawXML` capture helper implementing `xml.Marshaler`/`Unmarshaler` by replaying buffered tokens. Model WML with `xml:"URI local"` tags everywhere (never prefix-based tags), and make the blank document a set of hand-authored static XML byte blobs (styles/theme/fontTable/settings) assembled by the OPC writer in canonical order.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| ZIP container I/O, entry ordering | internal/opc | — | OPC is the container spec owner |
| [Content_Types].xml parse/serialize | internal/opc | — | Package-level manifest |
| Relationship graph (rIds, targets) | internal/opc | — | Spec lives in OPC Part 2, not WML |
| Safety limits (bomb, traversal, parts) | internal/opc | xmlutil (entity/DOCTYPE) | Container limits at ZIP layer; XML limits at decoder layer |
| Namespace URI↔prefix registry | internal/xmlutil | — | Shared by wml read/write; opc uses it only for rels/content-types (fixed ns) |
| WML type definitions | internal/wml | — | Schema ownership |
| RawXML hoarding mechanics | internal/xmlutil | internal/wml (fields) | xmlutil provides the type; wml embeds it |
| Blank default part content (styles, theme, fontTable, settings) | internal/wml (or root `wordingo` create path) | internal/opc (assembly) | Content is WML; packaging is OPC |
| Fixture corpus + golden diff harness | test infra (repo root `testdata/`) | — | Cross-cutting |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.23+ (floor; toolchain 1.26.5 verified installed) | Language/runtime | Locked D-10 |
| `archive/zip` | stdlib | OPC ZIP I/O | .docx is a ZIP; `NewReader(io.ReaderAt, size)` matches QUAL-01; `OpenRaw`/`Copy`/`CreateRaw` enable byte-exact pass-through [VERIFIED: pkg.go.dev/archive/zip go1.26.5] |
| `encoding/xml` | stdlib | XML decode/encode | URI-resolving unmarshal; token API for RawXML capture [VERIFIED: pkg.go.dev/encoding/xml behavior, see Code Examples] |
| `errors`, `fmt` | stdlib | Sentinel + wrapped errors | Locked D-11 |
| `testing` + golden files | stdlib | Fixture corpus tests | Locked D-01/D-04 |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `io` | stdlib | ReaderAt/Writer plumbing | QUAL-01 I/O flexibility |
| `strings`, `bytes` | stdlib | Part-name validation, buffer reuse | OPC path validation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Stdlib only | golang.org/x/net/xml | No meaningful advantage; violates zero-dep constraint |
| `encoding/xml` | Fully custom XML parser | Rejected explicitly by D-05 |

**Installation:** none — stdlib only.

**Version verification:** No external packages. Go toolchain verified: `go version go1.26.5 darwin/arm64` (this machine); go.mod floor `go 1.23`. All stdlib APIs used (`zip.OpenRaw`, `zip.Writer.Copy`, `zip.Writer.CreateRaw`, `zip.FileHeader.Modified`) exist since Go 1.17. [VERIFIED: pkg.go.dev/archive/zip]

## Package Legitimacy Audit

**Not applicable** — this phase installs zero external packages (stdlib-only constraint, PROJECT.md). No packages to audit, remove, or flag.

## Architecture Patterns

### System Architecture Diagram

```
Open path:
  io.ReaderAt ──► zip.NewReader ──► part enumeration (count ≤ maxParts)
                                      │
                                      ▼
                              [Content_Types].xml ──► ContentTypes map
                                      │
                              _rels/.rels ──► root Relationships
                                      │
                              per-part .rels ──► part Relationships
                                      │
                    ┌─────────────────┴──────────────────┐
                    ▼                                    ▼
            modeled parts (document,            unmodeled parts (customXml,
            styles, settings, theme,            glossary, media, VBA…)
            fontTable, numbering)               ──► raw []byte, lazily loaded
                    │                             via size-capped decompress
                    ▼                                    │
            xmlutil read-normalizer                      │
            (prefix→URI via encoding/xml)                │
                    │                                    │
                    ▼                                    ▼
            wml structs (+RawXML hoards) ──► Document model
                                                     │
Save path:                                             ▼
  Document model ──► wml marshal ──► xmlutil write-canonicalizer
                    (canonical w/r/a/wp prefixes, one xmlns set on root)
                    │
                    ▼
  merge {serialized modeled parts} ∪ {untouched raw parts}
                    │
                    ▼
  canonical entry order: [Content_Types].xml, _rels/.rels,
                         word/document.xml, word/_rels/document.xml.rels, …
                    │
                    ▼
  zip.Writer (fixed FileHeader.Modified for determinism;
              CreateRaw/Copy for untouched parts → byte-identical payload)
                    │
                    ▼
               io.Writer
```

### Recommended Project Structure

```
wordingo/
├── go.mod                    # module github.com/fabiomarini/wordingo; go 1.23
├── wordingo.go               # (plan 01-03) Create(); thin shell over internals
├── internal/
│   ├── opc/
│   │   ├── package.go        # Package open/save, part registry, lazy raw loading
│   │   ├── contenttypes.go   # [Content_Types].xml model (Defaults + Overrides)
│   │   ├── relationships.go  # Relationship sets, rId allocation, validation
│   │   ├── zipio.go          # canonical ordering, deterministic headers, size caps
│   │   └── errors.go         # sentinels (ErrInvalidPackage, ErrUnsafePath, …)
│   ├── xmlutil/
│   │   ├── ns.go             # URI↔prefix registry (canonical prefix table)
│   │   ├── encoder.go        # prefix-emitting token encoder (canonical write)
│   │   ├── rawxml.go         # RawXML token-blob capture + replay
│   │   └── decoder.go        # safe decoder factory (Strict, no DTD)
│   └── wml/
│       ├── document.go       # Document, Body, SectPr, P, R, T, …
│       ├── properties.go     # PPr, RPr and leaf property types
│       ├── styles.go         # Styles, Style, DocDefaults, LatentStyles, …
│       ├── numbering.go      # Numbering, AbstractNum, Num, Lvl, …
│       ├── table.go          # Tbl, Tr, Tc and properties
│       └── namespaces.go     # WML/relationship URI constants (Transitional+Strict)
└── testdata/
    ├── word/blank.docx, word/styled.docx
    ├── libreoffice/blank.docx, libreoffice/styled.docx
    ├── googledocs/blank.docx, googledocs/styled.docx
    └── hostile/customxml-glossary.docx          # (D-03)
```

### Pattern 1: URI-keyed struct tags (read side)
**What:** All WML struct tags use the `xml:"<namespace-URI> <local>"` form. Go's decoder resolves any prefix bound to that URI, satisfying OPC-03 on read with zero custom code.
**When to use:** Every WML struct field.
**Example:**
```go
// Source: pkg.go.dev/encoding/xml (Unmarshal: "a struct tag of the form
// 'name' or 'namespace-URL name' … the XML name … in the namespace")
type CT_P struct {
    XMLName xml.Name    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
    PPr     *CT_PPr     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
    R       []*CT_R     `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
    Raw     []RawXML    `xml:",any"`   // WML-04 hoard — see Pattern 3
}
```
**Caution:** on unmarshal into a field with a plain (non-URI) tag, Go matches only the local name — acceptable for attributes? No: `encoding/xml` does **not** resolve namespaces for attributes the same way; attribute tags must also use the `xml:"<URI> <local>,attr"` form for `w:val`-style namespaced attributes. `xml:space` uses the fixed URI `http://www.w3.org/XML/1998/namespace`. [CITED: pkg.go.dev/encoding/xml]

### Pattern 2: Canonical-prefix write encoder (the real xmlutil work)
**What:** `encoding/xml` marshals URI-tagged fields as `<local xmlns="URI">` — producing unprefixed elements with per-element xmlns re-declarations. Word accepts this but it is non-canonical, bloats output, and breaks per-part byte-diff goldens against producer fixtures. xmlutil's encoder walks the struct→token stream and rewrites `xml.Name{Space: URI}` to `prefix:local`, declaring all canonical `xmlns:*` on the root element once (registry per D-08: w, r, a, wp, wp14, mc, w14, w15, … plus `mc:Ignorable`).
**When to use:** Serializing any WML part.
**Example:**
```go
// xmlutil.Encoder — wraps xml.Encoder; before EncodeElement, replaces
// each token's Name.Space via Registry.PrefixFor(uri).
// Root element additionally receives synthesized xmlns:* attributes and
// mc:Ignorable listing ignorable extension prefixes (e.g. "w14 w15 wp14").
```

### Pattern 3: RawXML token-blob hoarding (D-06)
**What:** `RawXML` captures a complete element subtree (start→matching end) as buffered `xml.Token`s. Implements `xml.Marshaler`/`Unmarshaler`. On marshal, tokens are replayed verbatim (through the same prefix-canonicalizer, or raw for unknown namespaces whose prefixes are preserved from the source — recommend: replay raw, since unknown subtrees carry their own xmlns declarations).
**When to use:** `xml:",any"` catch-all fields on WML container structs (Body, P, R, PPr, RPr, Tbl, Tr, Tc, SectPr, Style, …).
**Example:**
```go
// xmlutil.RawXML
type RawXML struct{ Tokens []xml.Token }

func (r *RawXML) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
    r.Tokens = append(r.Tokens[:0], xml.CopyToken(start))
    depth := 1
    for depth > 0 {
        t, err := d.Token()
        if err != nil { return err }
        t = xml.CopyToken(t)
        switch t.(type) {
        case xml.StartElement: depth++
        case xml.EndElement:   depth--
        }
        r.Tokens = append(r.Tokens, t)
    }
    return nil
}
```
Note: `xml:",any"` into `[]RawXML` works only if RawXML implements Unmarshaler — it does above. Alternatively hoard with one `Raw []xml.Token` catch via a custom UnmarshalXML per container; the `[]RawXML` approach is simpler and matches D-06.

### Pattern 4: Canonical ZIP entry ordering (OPC-02)
**What:** ECMA-376 Part 2 §9.1.4.2 (streaming delivery) requires `[Content_Types].xml` be the first part in the package and `_rels/.rels` the second; Word tolerates other orders but several consumers don't, and PITFALLS.md flags it explicitly.
**Write order (fixed):**
1. `[Content_Types].xml`
2. `_rels/.rels`
3. `word/document.xml`
4. `word/_rels/document.xml.rels`
5. remaining `word/*.xml` parts (styles, settings, fontTable, theme, numbering…) each immediately followed by its `.rels` if present
6. everything else sorted by part name for determinism (media, customXml, docProps)
[CITED: ECMA-376 Part 2 §9.1.4.2 — via .planning/research/PITFALLS.md "Part Ordering"; exact full-spec text not re-fetched — LOW risk, golden tests against Word output will confirm]

**Determinism:** set `FileHeader.Modified` to a fixed time (e.g., 1980-01-01, the MS-DOS epoch; zero `Modified` already yields deterministic legacy fields) and never set `CreatorVersion`/extra fields — then whole-package bytes are reproducible, though D-04 only requires per-part diffs. [VERIFIED: pkg.go.dev/archive/zip FileHeader.Modified semantics]

### Pattern 5: Byte-exact pass-through (OPC-04)
**What:** Untouched parts are never decompressed/recompressed. Read: record `zip.File` + CRC; Save: `w.Copy(f)` (copies raw compressed form directly, bypassing decompression, compression, and validation) or `OpenRaw` + `CreateRaw`. Result: part bytes identical by construction — stronger than byte-diff equality after recompression.
**Constraint:** `Copy` preserves the original compressed stream; it does not let us rewrite the header (e.g., timestamps) — acceptable, D-04 diffs part payloads. Must skip `Copy` for parts whose name changed (none in Phase 1).
[VERIFIED: pkg.go.dev/archive/zip — Writer.Copy, File.OpenRaw, Writer.CreateRaw docs]

### Anti-Patterns to Avoid
- **Prefix-based struct tags (`xml:"w:p"`):** silently fails (empty structs, no error) on producers using other prefixes — the #1 encoding/xml/OOXML pitfall (PITFALLS.md Pitfall 3). Always URI tags.
- **Relying on default marshal output:** produces `<p xmlns="URI">` unprefixed soup; canonical-prefix encoder is mandatory for goldens.
- **Decompress-then-recompress pass-through:** Deflate output differs by level/implementation; breaks byte-identity (OPC-04). Use raw copy.
- **`xml:",innerxml"` for text capture:** trims/rewrites whitespace; use `,chardata` + explicit `xml:space` attribute handling (WML-03).
- **Merging adjacent runs on read:** breaks revision/field boundaries (WML-03). Never merge implicitly.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ZIP deflate I/O | custom inflate/deflate | `archive/zip` + raw copy | CRC, ZIP64, checksum edge cases |
| XML tokenization | custom scanner | `encoding/xml` Decoder (Strict) | entity/comment/CDATA/PI edge cases |
| XML entity expansion guard | hand parser | `encoding/xml` Strict decoder + reject `xml.Directive` (DOCTYPE) | encoding/xml does **not** expand internal DOCTYPE entities; unknown entities error out. Entity-expansion risk ≈ zero if DOCTYPE rejected [CITED: pkg.go.dev/encoding/xml — Decoder; ASSUMED strength of guarantee, test it] |
| Path traversal check | regex | `path/filepath.IsLocal`-style validation + OPC part-name rules | Go 1.20+ stdlib predicate matches `archive/zip`'s own `ErrInsecurePath` policy [VERIFIED: pkg.go.dev/archive/zip] |

**Key insight:** the only genuinely custom code is (a) canonical-prefix emission and (b) RawXML replay — both ~150–250 LOC each. Everything else is composition of stdlib primitives.

## Safety Limits (OPC-07) — Proposed Thresholds (agent's discretion per CONTEXT.md)

| Limit | Proposed Value | Basis |
|-------|---------------|-------|
| Max total decompressed size (all parts) | 512 MiB | PITFALLS.md security table recommendation [CITED: .planning/research/PITFALLS.md] |
| Max single-part decompressed size | 128 MiB | Word documents with large media; generous ceiling [ASSUMED] |
| Max compression ratio per entry | 100:1 (and flag warning at 20:1 for tiny compressed entries) | Zip-bomb detection heuristic; XML parts compress ~10–20:1 legitimately [ASSUMED] |
| Max parts | 4096 | PITFALLS.md suggests 1000 for ~400-page docs; 4096 gives headroom for glossary/customXml corpora [ASSUMED — CITED basis 1000] |
| Max XML nesting depth | 512 | Guards quadratic parser blowups; real WML rarely exceeds ~50 [ASSUMED] |
| Path validation | reject: absolute paths, `..` segments, backslashes, drive letters, empty segments, non-UTF-8 names; enforce part-name form `/word/document.xml` internally, ZIP-relative `word/document.xml` on disk | OPC Part 2 §8.1.1 part-name grammar + Go `ErrInsecurePath` policy [CITED: ECMA-376 Part 2; pkg.go.dev/archive/zip] |
| External relationships | parse and preserve, never fetch (SSRF guard); warn in `Warnings()` | PRD NFR security row [CITED: .planning/PRD.md §7] |

These are enforced with sentinels (`ErrDecompressionLimit`, `ErrUnsafePath`, `ErrTooManyParts`, `ErrXMLDepth`) wrapped per D-11.

## Conformance (OPC-05)

| Aspect | Transitional | Strict |
|--------|-------------|--------|
| WML main ns | `http://schemas.openxmlformats.org/wordprocessingml/2006/main` | `http://purl.oclc.org/ooxml/wordprocessingml/main` |
| Relationships ns (package) | `http://schemas.openxmlformats.org/package/2006/relationships` | `http://purl.oclc.org/ooxml/officeDocument/relationships` (careful: package-level rels ns stays `package/2006/relationships` in both; only officeDocument relationship URIs differ) |
| OfficeDocument rel URI | `…/officeDocument/2006/relationships/officeDocument` | `http://purl.oclc.org/ooxml/officeDocument/relationships/officeDocument` |

Policy: namespace registry maps **both** URI families to the same canonical prefixes (read-side URI-keyed tags must therefore tolerate both — implement wml decode with Transitional tags and a pre-pass that rewrites Strict URIs to Transitional in the token stream, or register both URIs per element; recommend token-stream URI normalization — one place, no tag duplication). Write: Transitional always for Create(); match source conformance when editing (record package conformance on open). [CITED: ECMA-376 Part 1 conformance classes, via PITFALLS.md Pitfall 6; exact Strict URI list — verify against fixture corpus if a Strict fixture appears; blank-doc path is Transitional-only so LOW risk]

## WML Type List (~60 proposed, 62 total)

Root/part types (7): `CT_Document` (w:document), `CT_Body`, `CT_Styles`, `CT_Numbering`, `CT_FontTable`/`CT_Fonts`, `CT_Settings`, `CT_Hdr`/`CT_Ftr` (shared block-content type `CT_HdrFtr`).

Paragraph/run core (10): `CT_P`, `CT_PPr`, `CT_R`, `CT_RPr`, `CT_Text` (w:t, with xml:space), `CT_Br`, `CT_Tab` (w:tab in run), `CT_Cr`, `CT_SoftHyphen`, `CT_NoBreakHyphen`.

Paragraph properties (13): `CT_PStyle` (w:pStyle), `CT_Jc`, `CT_Spacing` (w:spacing), `CT_Ind` (w:ind), `CT_NumPr` (w:numPr + w:numId + w:ilvl as 2 small types — count 3), `CT_KeepNext`, `CT_KeepLines`, `CT_PageBreakBefore`, `CT_WidowControl`, `CT_Tabs`/`CT_Tab` (tab stops — count 2), `CT_SectPr-ref` placeholder n/a. → subtotal 13 counting numPr sub-parts and tab stops.

Run properties (14): `CT_RStyle`, `CT_RFonts`, `CT_B` (w:b), `CT_I`, `CT_U` (w:underline), `CT_Sz`/`CT_SzCs` (1 type), `CT_Color`, `CT_Highlight`, `CT_VertAlign`, `CT_Lang`, `CT_SmallCaps`/`CT_Caps` (on/off type `CT_OnOff` shared — counts 1), `CT_Strike`/`CT_DStrike` (OnOff), `CT_Vanish`, `CT_Kern`. → 14.

Styles part (8): `CT_Style`, `CT_StyleName` (w:name), `CT_BasedOn`, `CT_Next`, `CT_Link`, `CT_DocDefaults`, `CT_RPrDefault`/`CT_PPrDefault` (1 wrapper type each — count 2), `CT_LatentStyles`/`CT_LsdException` (count 2, share file).

Numbering (6): `CT_Num`, `CT_AbstractNum`, `CT_Lvl`, `CT_NumFmt`, `CT_LvlText`, `CT_Start`.

Tables (10): `CT_Tbl`, `CT_TblPr`, `CT_TblStyle`, `CT_TblW`, `CT_TblBorders`(+ edge type `CT_TblBorder` count 1), `CT_TblGrid`/`CT_GridCol` (2), `CT_Tr`, `CT_TrPr`, `CT_Tc`, `CT_TcPr` (+ `CT_TcW` reuse `CT_TblW`, `CT_VMerge`, `CT_HMerge`-equivalent gridSpan `CT_GridSpan`, `CT_Shd` — count these 3 into TcPr line) → 10.

Sections (7): `CT_SectPr`, `CT_PgSz`, `CT_PgMar`, `CT_Cols`, `CT_DocGrid`, `CT_HdrFtrRef` (headerReference/footerReference), `CT_TitlePg` (OnOff).

Settings (4): `CT_Settings` root (above), `CT_Zoom`, `CT_DefaultTabStop`, `CT_Compat` (+`CT_CompatSetting`) — count 3 additional.

Theme (minimal, 3): `CT_Theme` opaque root holding `a:themeElements` — Phase 1 keeps theme as a **raw-modeled part** (typed shell + RawXML body) since full DrawingML is out of scope; types: `CT_Theme`, `a` handled as raw.

Shared simple types (folded into above): `CT_OnOff`, `CT_DecimalNumber`, hex/twips as Go `int64`/`string` fields with marshalers — not separate exported structs.

**Total ≈ 62 exported types** — within the "~60" bound. If count pressure arises, collapse OnOff leaf types (`CT_KeepNext`, `CT_Vanish`, …) into `*CT_OnOff` fields directly (they are all empty elements with optional `w:val`) — that removes ~8 types.

**Blank-document generator only needs ~25 of these** (document/body/p/r/t, pPr/rPr core, styles cluster, sectPr cluster, settings cluster). Tables/numbering types ship in Phase 1 (WML-01 requirement) but are exercised via marshal/unmarshal unit tests, not the blank doc.

## Blank Document Part Set (CREATE-01/02)

Minimal viable set Word opens without repair (matches python-docx/unioffice defaults and ECMA-376 §11 WordprocessingML part requirements):

| Part | Required | Content |
|------|----------|---------|
| `[Content_Types].xml` | yes | Defaults: `rels`, `xml`. Overrides: document (main), styles, settings, fontTable, theme |
| `_rels/.rels` | yes | one relationship → `word/document.xml` (officeDocument) |
| `word/document.xml` | yes | `<w:document><w:body><w:sectPr …/></w:body></w:document>`; sectPr with pgSz 12240×15840 (Letter), pgMar 1440/1440/1440/1440, cols, docGrid |
| `word/_rels/document.xml.rels` | yes | rels to styles, settings, fontTable, theme1 (rId1..4) |
| `word/styles.xml` | yes | docDefaults (Calibri 11pt / sz 22, theme fonts), latentStyles, Normal (default), Heading1–9 (basedOn Normal, next Normal, theme color accent1, outlineLvl 0–8), Title |
| `word/settings.xml` | yes | zoom 100, defaultTabStop 720, compat optional (omit — Word accepts absence) |
| `word/fontTable.xml` | yes | declarations for Calibri, Calibri Light, Cambria (or theme major/minor fonts) with panose/sig |
| `word/theme/theme1.xml` | yes | standard Office theme (a:theme with Office clrScheme/fontScheme/fmtScheme) — embed as static blob |
| `word/webSettings.xml` | **omit** | Optional; Word creates on demand. Omitting reduces surface [ASSUMED — but matches python-docx default template behavior] |
| `docProps/core.xml` + `app.xml` | **omit** | Optional; Word opens fine without them [ASSUMED — verify via golden/manual gate] |

Heading style essentials that Word's UI expects: each `w:style` for HeadingN carries `w:name val="heading N"`, `w:basedOn Normal`, `w:next Normal`, `w:qFormat`, `w:uiPriority`, pPr with `w:keepNext`, `w:spacing`, `w:outlineLvl`, rPr with `w:rFonts asciiTheme="majorHAnsi"`, theme color, size ramp (Heading1 sz 32 → Heading9 sz ~18/italic). Title: sz 52, thin bottom border optional (omit border to keep minimal). [CITED: ECMA-376 Part 1 §17 styles + Microsoft default template behavior; ASSUMED on exact sz ramp — copy values from a real Word blank fixture once D-02 fixtures land, or from python-docx `default.docx` styles.xml]

**Recommended construction:** embed `styles.xml`, `theme1.xml`, `fontTable.xml`, `settings.xml` as `//go:embed` static blobs authored once against a real Word blank document (diff-minimized), rather than generating them from WML structs. Guarantees CREATE-02 without depending on the correctness of 60 fresh marshal paths. document.xml + sectPr generated from wml structs (exercises the write path end-to-end).

## Common Pitfalls

### Pitfall 1: Prefix-based struct tags → silent empty structs
**What goes wrong:** `xml:"w:p"` matches only literal `w:` prefix; other producers' files parse into zero structs, no error.
**Avoid:** URI-form tags everywhere (Pattern 1). Test with fixtures from all 3 producers (D-03).
**Warning signs:** blank model after parse; styles "disappearing".

### Pitfall 2: Marshal emits `<p xmlns="…">` instead of `<w:p>`
**What goes wrong:** encoding/xml has no prefix registry; output is legal but non-canonical — golden diffs explode, file bloats, and some consumers (older Word validators) are pickier about root-level declarations than per-element redeclarations.
**Avoid:** Pattern 2 canonicalizer. Verify first golden against Word-authored blank.

### Pitfall 3: Whitespace collapse in w:t (WML-03)
**What goes wrong:** `,chardata` round-trip loses nothing, but re-serializing text without re-adding `xml:space="preserve"` when value has leading/trailing/double spaces corrupts rendering; Word *requires* the attribute to honor the spaces.
**Avoid:** `CT_Text{Value string; Space *string}` — on write set `xml:space="preserve"` iff `strings.TrimSpace(v) != v || strings.Contains(v, "  ")`; on read record verbatim. Never merge runs.
[CITED: PITFALLS.md Pitfall 8; ECMA-376 Part 1 §17.3.3.31 (w:t)]

### Pitfall 4: Recompression breaks OPC-04 byte-identity
**Avoid:** Pattern 5 raw copy. Also: do not "normalize" untouched XML parts (no re-indent, no re-decode).

### Pitfall 5: rId reuse/collision (OPC-06)
**Avoid:** allocator scans existing `.rels` once, then issues `rId<max+1..n>` monotonically per part; never reuse deleted ids; validate on save (every Target resolves, every part has content type, every r:embed reference has a rel). [CITED: PITFALLS.md Pitfall 9]

### Pitfall 6: `[Content_Types].xml` not first in archive
**Avoid:** Pattern 4 fixed write order. Note Go's `zip.Writer` streams in call order — control is trivial.

### Pitfall 7: `,any` catch-all silently dropping attributes of unknown elements
**What goes wrong:** RawXML replay must preserve attribute order and namespace declarations as captured — do not re-serialize StartElement attrs through a map (order loss).
**Avoid:** store tokens verbatim via `xml.CopyToken`; replay raw.

### Pitfall 8: Reading whole ZIP into memory / OOM on bombs
**Avoid:** streaming decompress via `io.LimitReader` wrappers around each `f.Open()` with per-part cap + running total cap (Safety Limits table). `zip.Reader` over `io.ReaderAt` already avoids full-buffer read of the archive itself.

## Code Examples

### Safe decoder factory (OPC-07 XML layer)
```go
// xmlutil/decoder.go
func NewSafeDecoder(r io.Reader) *xml.Decoder {
    d := xml.NewDecoder(io.LimitReader(r, maxPartXMLBytes))
    d.Strict = true          // unknown entities → error; encoding/xml never
                             // expands DOCTYPE-internal entities regardless
    d.Entity = nil           // no custom expansion
    return d
}
// Caller rejects xml.Directive tokens (DOCTYPE) at top of part parse:
//   if dir, ok := tok.(xml.Directive); ok { return ErrDOCTYPE }
// [CITED: pkg.go.dev/encoding/xml — Decoder.Strict, Directive; entity
//  non-expansion behavior — ASSUMED, verify with unit test using
//  billion-laughs fixture expecting a clean error]
```

### Byte-exact part copy on save (OPC-04)
```go
// internal/opc — untouched part pass-through
// [VERIFIED: pkg.go.dev/archive/zip Writer.Copy / File.OpenRaw]
for _, f := range zr.File {
    if modified[f.Name] || deleted[f.Name] { continue }
    if err := zw.Copy(f); err != nil { // raw form, no recompression
        return fmt.Errorf("opc: copy part %s: %w", f.Name, err)
    }
}
```
(Interleave with canonical-order emission: iterate the ordered name list, not `zr.File` order.)

### Deterministic part write (OPC-02)
```go
// [VERIFIED: pkg.go.dev/archive/zip FileHeader semantics]
h := &zip.FileHeader{Name: name, Method: zip.Deflate}
h.Modified = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC) // MS-DOS epoch
w, err := zw.CreateHeader(h)
```

### CT_Text with whitespace fidelity (WML-03)
```go
const xmlSpaceURI = "http://www.w3.org/XML/1998/namespace"

type CT_Text struct {
    XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
    Space   *string  `xml:"http://www.w3.org/XML/1998/namespace space,attr"`
    Value   string   `xml:",chardata"`
}
```

### Canonical prefix emission sketch (xmlutil write core)
```go
// Registry: uri → prefix (canonical, D-08)
var canonical = map[string]string{
    "http://schemas.openxmlformats.org/wordprocessingml/2006/main": "w",
    "http://schemas.openxmlformats.org/officeDocument/2006/relationships": "r",
    "http://schemas.openxmlformats.org/drawingml/2006/main": "a",
    "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing": "wp",
    // … w14, w15, wp14, mc, mc/AlternateContent …
}
// Encoder.EncodeToken: if Name.Space in registry → Name.Local = prefix+":"+local,
// Name.Space = "" — then delegate to xml.Encoder. Root token additionally
// receives synthesized Attr entries xmlns:w=… etc. once.
```

## LOC Budget Check (xmlutil)

| Component | Est. LOC |
|-----------|----------|
| ns registry + constants | ~80 |
| read normalizer (Strict-URI→Transitional rewrite, DOCTYPE reject, depth cap) | ~120 |
| write canonicalizer (prefix emitter + root xmlns synthesis + mc:Ignorable) | ~200 |
| RawXML capture/replay | ~80 |
| safe decoder + tests helpers | ~70 |
| **Total** | **~550** |

Within the ~500 LOC budget (PROJECT.md §11), below the 800 LOC revisit trigger. **Flag:** the mc:Ignorable/AlternateContent handling is the most likely LOC overrun source; if AlternateContent replay proves complex, keep AlternateContent subtrees inside RawXML (raw replay) and do not canonicalize their prefixes — fallback keeps budget.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| zip decompress→recompress round-trip | `zip.Writer.Copy` / `OpenRaw` raw pass-through | Go 1.17 | OPC-04 byte-identity becomes construction-level guarantee [VERIFIED: pkg.go.dev/archive/zip] |
| `FileHeader.SetModTime` (deprecated) | `FileHeader.Modified` field | Go 1.17 | Deterministic timestamps [VERIFIED] |
| Prefix-based xml tags | URI-based `xml:"URI local"` tags | Go 1.2 | Prefix-agnostic read (OPC-03) [VERIFIED] |

**Deprecated/outdated:**
- `FileHeader.ModifiedTime/ModifiedDate`, `SetModTime`, `ModTime()` — deprecated; use `Modified`. [VERIFIED: pkg.go.dev/archive/zip]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `encoding/xml` never expands DOCTYPE-internal entities; Strict mode errors on unknown entities | Don't Hand-Roll / Code Examples | If wrong, entity-expansion guard needs explicit entity-table rejection — add unit test with billion-laughs fixture in plan 01-01 |
| A2 | OPC-07 numeric thresholds (512MiB/128MiB/100:1/4096 parts/512 depth) | Safety Limits | User discretion item — confirm or adjust; low risk, values are generous |
| A3 | webSettings.xml and docProps may be omitted from blank doc | Blank Part Set | If Word 2016 complains (manual gate), add the parts — 30 min fix |
| A4 | Exact Heading 1–9 sz ramp and Title definition values | Blank Part Set | Cosmetic; copy real values from fixture once D-02 corpus lands |
| A5 | Strict-namespace URI normalization approach (token rewrite vs dual tags) | Conformance | Only matters if Strict fixtures exist; blank/create path unaffected |
| A6 | ECMA-376 §9.1.4.2 ordering details beyond "content-types first, .rels second" | Pattern 4 | Word is order-tolerant; golden tests confirm |

## Open Questions (RESOLVED)

1. (RESOLVED — 01-01 Task 0 fixture checkpoint, graceful skip) **Do D-02 fixtures exist yet?** Plan 01-01 tests depend on them.
   - What we know: D-02/D-03 commit the user to authoring 6–8 files under `testdata/`.
   - What's unclear: whether they're committed at plan-execution time.
   - Recommendation: first task of plan 01-01 is a checkpoint verifying fixture presence; unit tests generate synthetic fixtures meanwhile.
2. (RESOLVED — theme kept raw in Phase 1; CT_Theme RawXML shell in 01-02) **Theme part modeling depth.**
   - What we know: Phase 2 resolves theme colors (STYLE-RESOLVE-03).
   - What's unclear: whether Phase 1 should pre-parse `a:clrScheme` (6 elements) to ease Phase 2.
   - Recommendation: keep theme raw in Phase 1 (CREATE-01 only needs a valid blob); Phase 2 introduces DrawingML color types.
3. (RESOLVED — pass-through + prefix normalization, no special casing) **LibreOffice quirk tolerance on save.**
   - What we know: LibreOffice emits non-canonical prefixes and extra parts (`Configurations2`, `meta.xml` remnants in converted files).
   - Recommendation: pass-through handles extra parts; prefix normalization handles read — no special casing in Phase 1.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | everything | ✓ | go1.26.5 darwin/arm64 | — |
| go.mod floor 1.23 | D-10 | ✓ (1.26.5 ≥ 1.23) | — | — |
| External packages | none | n/a | — | — |
| Word 2016/2019/2021/M365 | CREATE-02 manual gate | ✗ (not on this machine, unverified) | — | manual user gate at sign-off (D-01) |
| LibreOffice | compatibility check | not probed | — | user-run |

**Missing dependencies with no fallback:** Word for the manual open gate — explicitly accepted as a user-performed sign-off step (D-01), not a blocker for implementation.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | OPC part-name grammar validation; ZIP entry policy; XML depth/size caps |
| V12 File Handling (files/resources) | yes | decompression-ratio + total-size limits; max parts; no auto-fetch of external relationships |
| V1 Architecture (attack surface) | yes | Strict XML decoder; DOCTYPE rejection; no DTD entity expansion |
| V2 Authentication / V3 Session / V4 Access Control | no | library, no auth surface |
| V6 Cryptography | no | none in Phase 1 |

### Known Threat Patterns for Go .docx library

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| ZIP bomb (42.zip style) | DoS | per-entry ratio cap + global decompressed-size cap via LimitReader |
| Billion laughs / entity expansion | DoS | encoding/xml (no entity expansion) + Strict + reject DOCTYPE Directive; unit-test billion-laughs input |
| Path traversal in part names / rel Targets | Tampering | reject `..`, absolute, backslash; validate Target resolution stays inside package |
| SSRF via external relationships (TargetMode="External") | Info disclosure | never resolve/fetch; warn in Warnings() (D-12) |
| Quadratic XML blowup (deep nesting) | DoS | token-depth cap (512) in xmlutil decoder |
| Huge part count (FD/memory exhaustion) | DoS | max-parts cap (4096) before parsing |

## Sources

### Primary (HIGH confidence)
- pkg.go.dev/archive/zip (go1.26.5 docs) — `Copy`, `OpenRaw`, `CreateRaw`, `FileHeader.Modified`, `ErrInsecurePath`, deprecated time APIs — fetched this session
- pkg.go.dev/encoding/xml — URI-form struct tags, `,any`, `,innerxml` whitespace caveat, Marshaler/Unmarshaler, Directive/Strict — cited from official docs
- .planning/research/PITFALLS.md — project-commissioned pitfalls research (namespace fragility, ordering §9.1.4.2, safety-limit baseline, whitespace) — HIGH within-project authority
- .planning/research/STACK.md, ARCHITECTURE.md — stack + layout guidance

### Secondary (MEDIUM confidence)
- ECMA-376 Part 2 §9.1.4.2 canonical ordering — via PITFALLS.md citation; spec text not directly re-fetched (c-rex mirror 404)
- ECMA-376 Part 1 style/heading structure — training knowledge consistent with PITFALLS.md; to be confirmed against fixture corpus (D-02)

### Tertiary (LOW confidence)
- Strict conformance purl.oclc.org URI list — normalize at read; validate only if a Strict fixture appears
- Heading sz ramp / Title values — placeholder until real fixture copy

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib APIs verified against live pkg.go.dev this session
- Architecture: HIGH — patterns anchored to verified stdlib capabilities + project pitfalls research
- Pitfalls: HIGH — sourced from commissioned PITFALLS.md + known encoding/xml behaviors
- Safety-limit numbers: MEDIUM — proposed values, user-discretion item
- Blank-doc part set: MEDIUM-HIGH — omit-optional-parts claims (A3) to be proven by manual Word gate

**Research date:** 2026-07-25
**Valid until:** 2026-08-24 (stable domain; stdlib semantics change slowly)
