package wordingo

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/style"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// FromTemplate opens a .docx template from path, clones its style
// dependency graph (styles, numbering, fontTable, theme, settings),
// and returns a Document with an empty body (one section, no
// paragraphs).  Style parts are never touched after CloneStyles
// (D-06).  Use when you want template styles but a clean body.
func FromTemplate(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wordingo: from template %s: %w", path, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("wordingo: stat template %s: %w", path, err)
	}
	return FromTemplateReader(f, fi.Size())
}

// FromTemplateReader is the io.ReaderAt variant of FromTemplate.
func FromTemplateReader(r io.ReaderAt, size int64) (*Document, error) {
	src, err := opc.Open(r, size)
	if err != nil {
		return nil, fmt.Errorf("wordingo: open template: %w", err)
	}

	dst := newTemplateTarget()
	if err := style.CloneStyles(src, dst); err != nil {
		return nil, fmt.Errorf("wordingo: clone styles: %w", err)
	}

	// FromTemplate clears body entirely — never modifies in place
	// (Pitfall 3: replace body, never edit existing one).
	dst.MarkModified("word/document.xml", buildEmptyBodyXML())

	doc, err := parseDocument(dst)
	if err != nil {
		return nil, err
	}
	return &Document{
		pkg:          dst,
		doc:          doc,
		nextImageID:  1,
		nextHeaderID: 1,
		nextFooterID: 1,
	}, nil
}

// OpenTemplate opens a .docx template from path, clones its style
// dependency graph, and returns a Document with the template's
// existing body content preserved (paragraphs and tables).  Header
// and footer references in sectPr are stripped (Pitfall 4 — cloned
// package has fresh rIds that would dangle).  Headers/footers
// themselves are not cloned in Phase 3 (deferred to Phase 5).
func OpenTemplate(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wordingo: open template %s: %w", path, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("wordingo: stat template %s: %w", path, err)
	}
	return OpenTemplateReader(f, fi.Size())
}

// OpenTemplateReader is the io.ReaderAt variant of OpenTemplate.
func OpenTemplateReader(r io.ReaderAt, size int64) (*Document, error) {
	src, err := opc.Open(r, size)
	if err != nil {
		return nil, fmt.Errorf("wordingo: open template: %w", err)
	}

	dst := newTemplateTarget()
	if err := style.CloneStyles(src, dst); err != nil {
		return nil, fmt.Errorf("wordingo: clone styles: %w", err)
	}

	// Read source body content, preserve paragraphs and tables,
	// replace sectPr with clean default (no HdrFtrRef/FtrRef — Pitfall 4).
	srcPart, ok := src.Parts["word/document.xml"]
	if !ok {
		return nil, fmt.Errorf("wordingo: template missing word/document.xml: %w", opc.ErrInvalidPackage)
	}
	rc, err := srcPart.Open()
	if err != nil {
		return nil, fmt.Errorf("wordingo: open template document.xml: %w", err)
	}
	defer rc.Close()

	var srcDoc wml.CT_Document
	dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
	if err := dec.Decode(&srcDoc); err != nil {
		return nil, fmt.Errorf("wordingo: decode template body: %w", err)
	}

	// Build fresh doc with source body content but clean sectPr.
	// This strips any header/footer references that would dangle
	// after CloneStyles allocates fresh rIds.
	freshDoc := &wml.CT_Document{
		Body: &wml.CT_Body{
			P:      srcDoc.Body.P,
			Tbl:    srcDoc.Body.Tbl,
			SectPr: defaultSectPr(),
		},
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(freshDoc); err != nil {
		return nil, fmt.Errorf("wordingo: encode template body: %w", err)
	}
	if err := enc.Flush(); err != nil {
		return nil, fmt.Errorf("wordingo: flush template body: %w", err)
	}
	dst.MarkModified("word/document.xml", buf.Bytes())

	doc, err := parseDocument(dst)
	if err != nil {
		return nil, err
	}
	return &Document{
		pkg:          dst,
		doc:          doc,
		nextImageID:  1,
		nextHeaderID: 1,
		nextFooterID: 1,
	}, nil
}
