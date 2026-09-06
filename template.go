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
// dependency graph (styles, numbering, fontTable, theme, settings) AND
// its letterhead — header/footer parts, their own relationship graphs,
// and the media (logos) they reference — and returns a Document with
// an empty body carrying the template's section properties.
//
// This is the "brand shell" semantics REFACTOR-HARNESS §3.3 requires:
// the look of a base template (styles + logo'd headers/footers + page
// geometry) with NONE of its example content. Style parts are never
// touched after CloneStyles (D-06).
//
// The Document's header/footer/image allocation counters start past
// the cloned parts, so AddHeader/AddImage on the result cannot
// overwrite the letterhead.
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

	// Read the source sectPr (letterhead references, page geometry).
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

	var srcSectPr *wml.CT_SectPr
	if srcDoc.Body != nil {
		srcSectPr = srcDoc.Body.SectPr
	}
	sectPr, counters := cloneLetterhead(src, dst, srcSectPr)

	// FromTemplate clears body content entirely — never modifies in
	// place (Pitfall 3: replace body, never edit existing one) — but
	// keeps the cloned letterhead's section properties.
	dst.MarkModified("word/document.xml", emptyBodyXMLWithSectPr(sectPr))

	doc, err := parseDocument(dst)
	if err != nil {
		return nil, err
	}
	d := &Document{
		pkg:          dst,
		doc:          doc,
		nextImageID:  counters.nextImage,
		nextHeaderID: counters.nextHeader,
		nextFooterID: counters.nextFooter,
	}
	d.syncBodyOrder()
	return d, nil
}

// OpenTemplate opens a .docx template from path, clones its style
// dependency graph, preserves the template's existing body content
// (paragraphs and tables), and clones the letterhead (headers/footers
// with their images) via the same path FromTemplate uses (D-13).
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

	// Read source body content.
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

	// Clone header/footer parts + their media (D-13).
	var srcSectPr *wml.CT_SectPr
	if srcDoc.Body != nil {
		srcSectPr = srcDoc.Body.SectPr
	}
	sectPr, counters := cloneLetterhead(src, dst, srcSectPr)

	var body *wml.CT_Body
	if srcDoc.Body != nil {
		body = &wml.CT_Body{
			P:      srcDoc.Body.P,
			Tbl:    srcDoc.Body.Tbl,
			SectPr: sectPr,
		}
	} else {
		body = &wml.CT_Body{SectPr: sectPr}
	}
	freshDoc := &wml.CT_Document{Body: body}

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
	d := &Document{
		pkg:          dst,
		doc:          doc,
		nextImageID:  counters.nextImage,
		nextHeaderID: counters.nextHeader,
		nextFooterID: counters.nextFooter,
	}
	d.syncBodyOrder()
	return d, nil
}

// findRelByID returns the relationship with the given ID from rs, or nil.
func findRelByID(rs *opc.Relationships, id string) *opc.Relationship {
	if rs == nil {
		return nil
	}
	for i := range rs.Rels {
		if rs.Rels[i].ID == id {
			return &rs.Rels[i]
		}
	}
	return nil
}

// readPartBytes reads the full payload of an OPC part.
func readPartBytes(part *opc.Part) ([]byte, error) {
	rc, err := part.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
