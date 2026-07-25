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
	"io"
	"os"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
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
	pkg *opc.Package
	doc *wml.CT_Document
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
	return &Document{pkg: pkg, doc: doc}, nil
}

// WriteTo writes the document to w. Returns bytes written.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
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

// Warnings returns non-fatal issues from the package layer.
func (d *Document) Warnings() []string {
	return d.pkg.Warnings()
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
		paras[i] = &Paragraph{ct: p}
	}
	return paras
}

// Close releases package and document references. The Document is
// not usable after Close.
func (d *Document) Close() error {
	d.pkg = nil
	d.doc = nil
	return nil
}
