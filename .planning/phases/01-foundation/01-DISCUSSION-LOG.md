# Phase 1: Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-07-25
**Phase:** 01-foundation
**Areas discussed:** Word-compat verification, XML layer strategy, Module & repo identity, Error/warning API shape

---

## Word-Compat Verification

| Option | Description | Selected |
|--------|-------------|----------|
| Fixture corpus + golden files | Commit .docx fixtures from Word/LibreOffice/Google Docs + golden files; manual Word-open check before sign-off | ✓ |
| OpenXml SDK validator in CI | .NET container validates generated files | |
| Both combined | Corpus + validator container | |

| Option | Description | Selected |
|--------|-------------|----------|
| Author fixtures now, commit | User creates fixtures in Word + LibreOffice + Google Docs under testdata/ | ✓ |
| Self-generated fixtures | Synthetic fixtures from own writer | |
| You decide | Researcher picks strategy | |

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal 6-8 fixture set | blank.docx + styled.docx per producer + one hostile file (customXml, glossary) | ✓ |
| Broad corpus | Many varied real documents | |
| You decide | | |

| Option | Description | Selected |
|--------|-------------|----------|
| Per-part byte diff | Unzip both, diff each part | ✓ |
| Whole-file identity | Whole-ZIP byte identity | |
| You decide | | |

**Notes:** Validator container declined; revisit if golden files prove insufficient.

---

## XML Layer Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Token-stream wrapper | xmlutil normalizes prefixes→URIs over encoding/xml; structs keep xml tags | ✓ |
| Custom parser | Fully hand-rolled tokenizer/parser | |
| You decide | Researcher evaluates both | |

| Option | Description | Selected |
|--------|-------------|----------|
| Raw subtree blobs | Unknown children as xmlutil.RawXML blobs, re-emitted verbatim | ✓ |
| Generic element tree | Name/attrs/children tree for unknown parts | |
| You decide | | |

| Option | Description | Selected |
|--------|-------------|----------|
| Pointers for optional | nil = absent | ✓ |
| Presence flags | Value fields + flags | |
| You decide | | |

| Option | Description | Selected |
|--------|-------------|----------|
| Canonical OOXML prefixes | Emit w, r, a, wp on write; read any prefix by URI | ✓ |
| Source-echo prefixes | Preserve each file's own prefixes | |
| You decide | | |

---

## Module & Repo Identity

| Option | Description | Selected |
|--------|-------------|----------|
| personal GitHub path | github.com/fabiomarini/wordingo (matches PRD) | ✓ |
| Other path | Org or custom domain | |

| Option | Description | Selected |
|--------|-------------|----------|
| 1.23 floor | go 1.23 in go.mod, CI tests 1.23 + latest | ✓ |
| Latest-2 policy | Track latest two releases | |
| You decide | | |

---

## Error/Warning API Shape

| Option | Description | Selected |
|--------|-------------|----------|
| Sentinels + wrapping | ErrXxx sentinels + fmt.Errorf %w; errors.Is | ✓ |
| Typed error structs | Structured error fields | |
| You decide | | |

| Option | Description | Selected |
|--------|-------------|----------|
| Plain strings | Warnings() []string | ✓ |
| Typed warnings | []Warning{Code, Message} | |
| You decide | | |

---

## the agent's Discretion

Internal package organization details, exact ~60 WML type list, test conventions, OPC-07 numeric thresholds.

## Deferred Ideas

- Open XML SDK validator in CI — revisit if golden-file approach insufficient.
- V2 features per REQUIREMENTS.md — behind v1 validation.
