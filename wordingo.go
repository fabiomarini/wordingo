// Package wordingo creates and edits WordprocessingML (.docx)
// documents. It is a pure Go, zero-dependency library built on
// github.com/fabiomarini/wordingo/internal/opc and
// internal/{wml,xmlutil}.
//
// Create a blank document:
//
//	doc, err := wordingo.Create()
//	if err != nil { ... }
//	defer doc.Close()
//	err = doc.Save("output.docx")
//
// Open an existing document:
//
//	doc, err := wordingo.Open("existing.docx")
//	if err != nil { ... }
//	defer doc.Close()
//	for _, p := range doc.Paragraphs() { ... }
//	err = doc.Save("roundtrip.docx")
package wordingo

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// countWriter wraps io.Writer and counts bytes written.
type countWriter struct {
	w io.Writer
	n int64
}

func (cw *countWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}

// Document wraps an OPC package during construction.  Use Create to
// create a blank document, Open or OpenReader to open an existing
// document, then Save or WriteTo to persist.  Close releases internal
// references.
type Document struct {
	pkg           *opc.Package
	doc           *wml.CT_Document
	dirty         bool
	warnings      []string
	nextImageID   int64
	nextHeaderID  int64
	nextFooterID  int64
}

// Create returns a new blank document with default styles (Normal,
// Heading 1–9, Title), theme, font table, settings, and one section
// page sized Letter (12240×15840 twips) with 1-inch margins.
func Create() (*Document, error) {
	pkg := newBlankPackage()
	doc, err := parseDocument(pkg)
	if err != nil {
		return nil, err
	}
	d := &Document{
		pkg:          pkg,
		doc:          doc,
		nextImageID:  1,
		nextHeaderID: 1,
		nextFooterID: 1,
	}
	d.syncBodyOrder()
	return d, nil
}

// WriteTo writes the document to w. Returns bytes written.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
	if d == nil {
		panic("wordingo: WriteTo called on nil Document")
	}
	d.serializeBody()
	cw := &countWriter{w: w}
	err := d.pkg.Save(cw)
	return cw.n, err
}

// Save writes the document to a file at path.
func (d *Document) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = d.WriteTo(f)
	return err
}

// SaveFile writes the document to a file at path.
func (d *Document) SaveFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.pkg.Save(f)
}

// Warnings returns non-fatal issues from the package layer and
// document-level formatting validation.
func (d *Document) Warnings() []string {
	if d == nil {
		panic("wordingo: Warnings called on nil Document")
	}
	var all []string
	all = append(all, d.pkg.Warnings()...)
	all = append(all, d.warnings...)
	return all
}

// warn appends a formatted warning to d.warnings.
func (d *Document) warn(format string, args ...any) {
	d.warnings = append(d.warnings, fmt.Sprintf(format, args...))
}

// checkStyleNames validates all style references in the body against
// available style IDs in word/styles.xml. Appends warnings for unknown
// or XML-illegal style names.
func (d *Document) checkStyleNames() {
	known := make(map[string]bool)
	if part, ok := d.pkg.Parts["word/styles.xml"]; ok {
		rc, err := part.Open()
		if err == nil {
			var styles wml.CT_Styles
			dec := xmlutil.NewSafeDecoder(rc, opc.MaxPartBytes)
			if err := dec.Decode(&styles); err == nil {
				for _, s := range styles.Style {
					if s.StyleID != nil && *s.StyleID != "" {
						known[*s.StyleID] = true
					}
				}
			}
			rc.Close()
		}
	}

	hasXMLIllegal := func(s string) bool {
		for _, c := range s {
			if c == 0 || (c < 0x20 && c != '\t' && c != '\n' && c != '\r') {
				return true
			}
		}
		return false
	}

	checkName := func(name string, context string) {
		if name == "" {
			return
		}
		if hasXMLIllegal(name) {
			d.warn("wordingo: style %q in %s contains XML-illegal characters", name, context)
			return
		}
		if !known[name] {
			d.warn("wordingo: unknown style %q referenced by %s", name, context)
		}
	}

	if d.doc == nil || d.doc.Body == nil {
		return
	}
	for pi, para := range d.doc.Body.P {
		if para.PPr != nil && para.PPr.PStyle != nil && para.PPr.PStyle.Val != nil {
			checkName(*para.PPr.PStyle.Val, fmt.Sprintf("paragraph %d", pi))
		}
		for ri, run := range para.R {
			if run.RPr != nil && run.RPr.RStyle != nil && run.RPr.RStyle.Val != nil {
				checkName(*run.RPr.RStyle.Val, fmt.Sprintf("paragraph %d run %d", pi, ri))
			}
		}
	}
}

// serializeBody re-encodes the document body to XML if the dirty
// flag is set, then calls MarkModified so changes are persisted at
// Save time.
func (d *Document) serializeBody() {
	if d == nil {
		panic("wordingo: serializeBody called on nil Document")
	}
	if !d.dirty {
		return
	}
	d.checkStyleNames()
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(d.doc); err != nil {
		d.warn("wordingo: serialize body: %v", err)
		return
	}
	if err := enc.Flush(); err != nil {
		d.warn("wordingo: flush body: %v", err)
		return
	}
	d.pkg.MarkModified("word/document.xml", buf.Bytes())
	d.dirty = false
}

// X returns the underlying OPC package for escape-hatch access to
// internals (content types, relationships, individual parts).
func (d *Document) X() *opc.Package {
	return d.pkg
}

// Paragraphs returns the document's body paragraphs as read-only
// wrappers over *wml.CT_P. Returns an empty slice if the document
// body is nil.
func (d *Document) Paragraphs() []*Paragraph {
	if d.doc == nil || d.doc.Body == nil {
		return nil
	}
	paras := make([]*Paragraph, len(d.doc.Body.P))
	for i, p := range d.doc.Body.P {
		paras[i] = &Paragraph{ct: p, doc: d}
	}
	return paras
}

// AddParagraph appends a paragraph with optional text and returns it.
func (d *Document) AddParagraph(text string) *Paragraph {
	if d == nil {
		panic("wordingo: AddParagraph called on nil Document")
	}
	ct := &wml.CT_P{}
	if text != "" {
		ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
	}
	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	d.doc.Body.AppendP(ct)
	d.dirty = true
	return &Paragraph{ct: ct, doc: d}
}

// InsertBefore inserts a new paragraph with the given text before target
// (identified by pointer identity). Returns the new paragraph, or nil if
// target is not found in body paragraphs.
func (d *Document) InsertBefore(target *Paragraph, text string) *Paragraph {
	if d == nil {
		panic("wordingo: InsertBefore called on nil Document")
	}
	if target == nil {
		d.warn("wordingo: InsertBefore: target is nil")
		return nil
	}
	ct := &wml.CT_P{}
	if text != "" {
		ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
	}
	for i, p := range d.doc.Body.P {
		if p == target.ct {
			d.doc.Body.P = append(d.doc.Body.P, nil)
			copy(d.doc.Body.P[i+1:], d.doc.Body.P[i:])
			d.doc.Body.P[i] = ct
			d.insertBodyOrderAtPIndex(i)
			d.dirty = true
			return &Paragraph{ct: ct, doc: d}
		}
	}
	d.warn("wordingo: InsertBefore: target paragraph not found")
	return nil
}

// InsertAfter inserts a new paragraph with the given text after target
// (identified by pointer identity). Returns the new paragraph, or nil if
// target is not found in body paragraphs.
func (d *Document) InsertAfter(target *Paragraph, text string) *Paragraph {
	if d == nil {
		panic("wordingo: InsertAfter called on nil Document")
	}
	if target == nil {
		d.warn("wordingo: InsertAfter: target is nil")
		return nil
	}
	ct := &wml.CT_P{}
	if text != "" {
		ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
	}
	for i, p := range d.doc.Body.P {
		if p == target.ct {
			d.doc.Body.P = append(d.doc.Body.P, nil)
			copy(d.doc.Body.P[i+2:], d.doc.Body.P[i+1:])
			d.doc.Body.P[i+1] = ct
			d.insertBodyOrderAtPIndex(i + 1)
			d.dirty = true
			return &Paragraph{ct: ct, doc: d}
		}
	}
	d.warn("wordingo: InsertAfter: target paragraph not found")
	return nil
}

// DeleteParagraph removes target paragraph (identified by pointer identity)
// from the document body. If target is not found, warns and returns.
func (d *Document) DeleteParagraph(target *Paragraph) {
	if d == nil {
		panic("wordingo: DeleteParagraph called on nil Document")
	}
	if target == nil {
		d.warn("wordingo: DeleteParagraph: target is nil")
		return
	}
	for i, p := range d.doc.Body.P {
		if p == target.ct {
			d.doc.Body.P = append(d.doc.Body.P[:i], d.doc.Body.P[i+1:]...)
			d.deleteBodyOrderAtPIndex(i)
			d.dirty = true
			return
		}
	}
	d.warn("wordingo: DeleteParagraph: target paragraph not found")
}

// Close releases package and document references. The Document is
// not usable after Close.
func (d *Document) Close() error {
	d.pkg = nil
	d.doc = nil
	return nil
}
