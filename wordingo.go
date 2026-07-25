// Package wordingo creates and edits WordprocessingML (.docx)
// documents. It is a pure Go, zero-dependency library built on
// github.com/fabiomarini/wordingo/internal/opc and
// internal/{wml,xmlutil}.
//
// Create a blank document:
//
//	doc := wordingo.Create()
//	err := doc.SaveFile("output.docx")
//	// or: doc.Save(w)
package wordingo

import (
	"io"
	"os"

	"github.com/fabiomarini/wordingo/internal/opc"
)

// Document wraps an OPC package during construction.  Use Create to
// create a blank document, then Save or SaveFile to write it out.
type Document struct {
	pkg *opc.Package
}

// Create returns a new blank document with default styles (Normal,
// Heading 1–9, Title), theme, font table, settings, and one section
// page sized Letter (12240×15840 twips) with 1-inch margins.
func Create() *Document {
	return &Document{pkg: newBlankPackage()}
}

// Save writes the document to w.
func (d *Document) Save(w io.Writer) error {
	return d.pkg.Save(w)
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
