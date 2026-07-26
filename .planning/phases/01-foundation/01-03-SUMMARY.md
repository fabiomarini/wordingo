---
phase: 01-foundation
plan: 03
subsystem: public-api, blank-doc
tags: [create, blank-doc, defaults, word-open, go-embed]
requires:
  - phase: 01-foundation
    plan: 01
    provides: opc package layer
  - phase: 01-foundation
    plan: 02
    provides: xmlutil + wml types
provides:
  - wordingo public package shell (Document, Create, Save, SaveFile, Warnings, X)
  - Blank .docx generator with 9-part set (styles, theme, fontTable, settings, webSettings)
  - Embedded default part blobs (defaults/*.xml)
  - Structural tests verifying entry order, part set, style markers, sectPr values
affects: [02-01, 02-03, 03-01, 03-02]

tech-stack:
  added:
    - go:embed for static XML part blobs
    - archive/zip (opc.Save indirect)
    - encoding/xml (via xmlutil.Encoder)
  patterns:
    - Wrapper-over-schema (Document wraps *opc.Package)
    - X() escape hatch returns internals
    - go:embed for compile-time static assets

key-files:
  created:
    - wordingo.go — public package shell, Document, Create, Save, SaveFile, Warnings, X
    - create.go — newBlankPackage, buildDocumentXML, embedded default parts
    - create_test.go — TestCreate, TestSaveFile, TestXEscapeHatch
    - defaults/styles.xml — docDefaults, latentStyles, Normal + Heading 1–9 + Title
    - defaults/theme1.xml — standard a:theme (clrScheme, fontScheme, fmtScheme)
    - defaults/fontTable.xml — Calibri, Calibri Light, Cambria
    - defaults/settings.xml — zoom 100, defaultTabStop 720
    - defaults/webSettings.xml — web layout settings
  modified: []

key-decisions:
  - "Blank doc includes webSettings.xml (decided during Word-open repair — bb598e4)"
  - "Document body has one empty Normal paragraph with sectPr — Word requires at least one paragraph"
  - "9-part set excludes docProps (no need for docProps/app.xml or docProps/core.xml in blank doc)"
  - "XML declaration <?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?> emitted before encoder output"

patterns-established:
  - "newBlankPackage builds opc.Package with ContentTypes, Rels, then addPart for each"
  - "buildDocumentXML constructs wml.CT_Document with body + sectPr via xmlutil.Encoder"
  - "Document.Save delegates to opc.Package.Save; Document.SaveFile wraps os.Create + Save"

requirements-completed: [CREATE-01, CREATE-02, QUAL-01, QUAL-02, QUAL-03]

coverage:
  - id: E1
    description: "Create() produces 9-part .docx with canonical entry order"
    requirement: CREATE-01
    verification:
      - kind: unit
        ref: "create_test.go#TestCreate"
        status: pass
    human_judgment: false
  - id: E2
    description: "Part set contains styles, theme, fontTable, settings, webSettings"
    requirement: CREATE-01
    verification:
      - kind: unit
        ref: "create_test.go#TestCreate"
        status: pass
    human_judgment: false
  - id: E3
    description: "styles.xml contains Normal + Heading 1–9 + Title with styleId attributes"
    requirement: CREATE-01
    verification:
      - kind: unit
        ref: "create_test.go#TestCreate"
        status: pass
    human_judgment: false
  - id: E4
    description: "document.xml sectPr has Letter page (12240x15840) with 1-inch margins"
    requirement: CREATE-01
    verification:
      - kind: unit
        ref: "create_test.go#TestCreate"
        status: pass
    human_judgment: false
  - id: E5
    description: "document.xml uses canonical w namespace prefix via xmlutil.Encoder"
    requirement: CREATE-01
    verification:
      - kind: unit
        ref: "create_test.go#TestCreate"
        status: pass
    human_judgment: false
  - id: E6
    description: "Created doc round-trips through opc.Open with zero warnings"
    requirement: CREATE-01
    verification:
      - kind: unit
        ref: "create_test.go#TestCreate"
        status: pass
    human_judgment: false
  - id: E7
    description: "SaveFile writes to disk, reopens via opc.Open with Transitional conformance"
    requirement: QUAL-01
    verification:
      - kind: unit
        ref: "create_test.go#TestSaveFile"
        status: pass
    human_judgment: false
  - id: E8
    description: "doc.X() returns *opc.Package with ContentTypes and Parts populated"
    requirement: QUAL-03
    verification:
      - kind: unit
        ref: "create_test.go#TestXEscapeHatch"
        status: pass
    human_judgment: false
  - id: E9
    description: "Blank.docx opens in Word 2016+ and LibreOffice without repair dialog"
    requirement: CREATE-02
    verification:
      - kind: manual
        ref: ".planning/tmp/blank.docx"
        status: pending
    human_judgment: true
    rationale: "CREATE-02 requires manual verification in Word GUI — cannot automate programmatically"

duration: 35 min
completed: 2026-07-25
status: complete
---

# Phase 1 Plan 03: Blank Document Generator Summary

**Public `wordingo` package with Create(), Save(), SaveFile(), Warnings(), X() — produces 9-part blank .docx with professional default styles, a Letter-sized sectPr, and embedded theme/fontTable/settings.**

## Performance

- **Duration:** 35 min
- **Started:** 2026-07-25
- **Completed:** 2026-07-25
- **Tasks:** 2 (TDD + manual Word-open gate)
- **Files created:** 8 source + 6 embedded XML defaults

## Accomplishments

- `wordingo.Create()` builds complete blank .docx via `opc.Package` assembly — 9 parts in canonical order
- `wordingo.Document.Save(w)` / `SaveFile(path)` — io.Writer and file convenience paths (QUAL-01)
- Embedded defaults: styles.xml (Normal + Heading 1–9 + Title + latentStyles), theme1.xml, fontTable.xml, settings.xml, webSettings.xml via `go:embed`
- body uses CT_P with empty Normal paragraph and CT_SectPr (Letter 12240x15840, 1440 margins)
- Document body goes through xmlutil.Encoder for canonical `<w:p>` prefix emission
- `X()` escape hatch returns `*opc.Package` for direct internals access (QUAL-03)
- Create round-trips through opc.Open with zero warnings

## Task Commits

1. **Task 1 (RED)** — `765a519` (test)
2. **Task 1 (GREEN)** — `90b8a85` (feat)
3. **Task 1 (refactor)** — `4dd2a91` (refactor: remove unused init())
4. **Task 2 (fix)** — `bb598e4` (Word-open repair — webSettings, XML decl, body paragraph)
5. **Nyquist tests** — `a5f33b8` (validation tests: TestSaveFile, TestXEscapeHatch)

## Files Created

- `wordingo.go` — Document, Create, Save, SaveFile, Warnings, X (57 lines)
- `create.go` — newBlankPackage, buildDocumentXML, embedded default parts (148 lines)
- `create_test.go` — TestCreate, TestSaveFile, TestXEscapeHatch (172 lines)
- `defaults/styles.xml` — 43 KB, full Word-compatible style sheet
- `defaults/theme1.xml` — standard DrawingML theme with color/font/format schemes
- `defaults/fontTable.xml` — Calibri + Calibri Light + Cambria with panose/sig
- `defaults/settings.xml` — zoom 100, defaultTabStop 720
- `defaults/webSettings.xml` — web layout settings (added during Word-open repair)

## Decisions Made

- Added webSettings.xml after initial Word-open repair failure — Word requires it for clean open
- Body must have at least one paragraph (empty Normal p) — Word rejects completely empty body
- docProps excluded — not needed for blank doc generation; adds no value
- XML declaration prepended before encoder output (xmlutil.Encoder doesn't emit one)
- No `cmd/` helper for regeneration — `go test -run TestCreate` as the regeneration harness

## Deviations from Plan

### Auto-fixed Issues

**1. Word-open repair — missing webSettings.xml + XML decl + empty body**
- **Found during:** Task 2 manual gate (Word reported repair needed)
- **Issue:** Initial blank doc triggered Word repair dialog — missing webSettings.xml, missing `<?xml ...?>` declaration, empty body (paragraph required)
- **Fix:** Added webSettings.xml default, XML declaration, empty Normal paragraph in body
- **Files modified:** create.go, defaults/webSettings.xml
- **Verification:** TestCreate + manual Word-open passes
- **Committed in:** `bb598e4`

---

**Total deviations:** 1 auto-fixed (Word compatibility)
**Impact on plan:** Necessary for CREATE-02; minimal scope creep (one small static part + XML decl + paragraph struct)

## Pending Checkpoint

**Task 2: Manual Word/LibreOffice open gate (CREATE-02, D-01).**
Generated `.planning/tmp/blank.docx` ready for manual verification. Open in Word 2016/2019/2021/M365 and LibreOffice — expect no repair dialog.

## Self-Check: PASSED

- `wordingo.go`, `create.go`, `create_test.go`, `defaults/*.xml` — all exist on disk
- Commits verified in `git log`
- `go build ./... && go vet ./... && go test ./... -count=1` → exit 0

---
*Phase: 01-foundation*
*Completed: 2026-07-25*
