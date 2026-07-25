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

// Open reads an existing .docx file from path and returns a Document
// with the body parsed eagerly. Supporting parts remain lazy.
func Open(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wordingo: open %s: %w", path, err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("wordingo: stat %s: %w", path, err)
	}
	return OpenReader(f, fi.Size())
}

// OpenReader reads a .docx from r with the given size and returns a
// Document with the body parsed eagerly. Supporting parts remain lazy.
func OpenReader(r io.ReaderAt, size int64) (*Document, error) {
	pkg, err := opc.Open(r, size)
	if err != nil {
		return nil, fmt.Errorf("wordingo: %w", err)
	}
	doc, err := parseDocument(pkg)
	if err != nil {
		return nil, err
	}
	return &Document{pkg: pkg, doc: doc}, nil
}

// parseDocument reads word/document.xml from the package and
// unmarshals it into *wml.CT_Document.
func parseDocument(pkg *opc.Package) (*wml.CT_Document, error) {
	part, ok := pkg.Parts["word/document.xml"]
	if !ok {
		return nil, fmt.Errorf("wordingo: missing word/document.xml: %w", opc.ErrInvalidPackage)
	}
	rc, err := part.Open()
	if err != nil {
		return nil, fmt.Errorf("wordingo: open document.xml: %w", err)
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return nil, fmt.Errorf("wordingo: read document.xml: %w", err)
	}

	dec := xmlutil.NewSafeDecoder(bytes.NewReader(buf.Bytes()), opc.MaxPartBytes)
	var doc wml.CT_Document
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("wordingo: decode document.xml: %w", err)
	}
	return &doc, nil
}
