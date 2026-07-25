---
phase: 03-document-model
plan: 02
subsystem: wordingo
tags:
  - template
  - clone-styles
  - body-preservation
  - roundtrip
requires:
  - 03-01
provides:
  - FromTemplate
  - FromTemplateReader
  - OpenTemplate
  - OpenTemplateReader
  - buildEmptyBodyXML
affects:
  - create.go (newTemplateTarget, ptrInt64, defaultSectPr, buildEmptyBodyXML)
  - roundtrip_test.go (removed duplicate defaultSectPr)
tech-stack:
  added: []
  patterns:
    - template cloning via CloneStyles byte pass-through
    - synthetic fixture builder for template tests
key-files:
  created:
    - template.go
    - template_test.go
  modified:
    - create.go
    - roundtrip_test.go
key-decisions: []
requirements-completed:
  - CREATE-03
  - CREATE-04
  - STYLE-ROUNDTRIP-01
duration: "24 min"
completed: "2026-07-25T23:30:00Z"
status: complete
coverage:
  - deliverable: FromTemplate/FromTemplateReader
    verification:
      - kind: test
        ref: TestFromTemplate_EmptyBody
        status: pass
      - kind: test
        ref: TestFromTemplate_RoundTrip
        status: pass
    human_judgment: false
  - deliverable: OpenTemplate/OpenTemplateReader
    verification:
      - kind: test
        ref: TestOpenTemplate_BodyPreserved
        status: pass
      - kind: test
        ref: TestOpenTemplate_RoundTrip
        status: pass
    human_judgment: false
  - deliverable: buildEmptyBodyXML
    verification:
      - kind: test
        ref: TestBuildEmptyBodyXML_Valid
        status: pass
    human_judgment: false
  - deliverable: Path-based convenience functions
    verification:
      - kind: test
        ref: TestFromTemplate_And_OpenTemplate_ReaderVariants
        status: pass
    human_judgment: false
  - deliverable: Template from blank doc (edge case)
    verification:
      - kind: test
        ref: TestTemplateFromBlankDoc
        status: pass
    human_judgment: false
---

# Phase 3 Plan 2: FromTemplate/OpenTemplate Summary

FromTemplate and OpenTemplate — template document creation with style cloning via Phase 2's CloneStyles and configurable body policy (clear vs preserve). buildEmptyBodyXML helper for clean section-provisioned empty bodies with no dangling header/footer references.

## Files Created/Modified

| File | Status | Purpose |
|------|--------|---------|
| `template.go` | Created | FromTemplate, FromTemplateReader, OpenTemplate, OpenTemplateReader |
| `template_test.go` | Created | All 7 template tests |
| `create.go` | Modified | Added ptrInt64, defaultSectPr, buildEmptyBodyXML, newTemplateTarget |
| `roundtrip_test.go` | Modified | Removed duplicate defaultSectPr (moved to create.go) |

## Key Implementation Details

- **FromTemplate**: Clones 5 style parts via CloneStyles (byte pass-through), replaces body with buildEmptyBodyXML (no paragraphs, Letter page, 1in margins). Body is replaced entirely — not modified in place (Pitfall 3).
- **OpenTemplate**: Same style clone, keeps body paragraphs and tables from source, replaces sectPr with clean default (no HdrFtrRef/FtrRef — Pitfall 4 avoidance).
- **newTemplateTarget()**: Minimal package without style parts, satisfying CloneStyles' fresh-empty-target precondition (D-08). Unlike newBlankPackage(), it has only infrastructure parts.
- **buildEmptyBodyXML()**: Reuses defaultSectPr(), encodes via xmlutil.NewEncoder matching buildDocumentXML pattern.
- **No MarkModified on style parts** from template code — only CloneStyles touches them (D-06).

## Deviations from Plan

No deviations — one adaption: `newBlankPackage()` already creates style parts, which conflicts with CloneStyles' fresh-empty-target precondition. Created `newTemplateTarget()` as a minimal alternative without style parts. This is an internal implementation detail, not an API change.

## Verification Results

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test . -run 'TestFromTemplate\|TestOpenTemplate\|TestTemplate' -v -count=1` | 7/7 PASS |
| `go test . -run 'TestRoundTrip' -v -count=1` | 13/13 PASS |
| `go vet .` | PASS |

## Self-Check

```
FOUND: template.go
FOUND: template_test.go
FOUND: create.go (modified)
FOUND: roundtrip_test.go (modified)
FOUND: 1eb836d (feat: implement FromTemplate, OpenTemplate, buildEmptyBodyXML)
FOUND: f40d42d (test: add template tests)
```

## Self-Check: PASSED
