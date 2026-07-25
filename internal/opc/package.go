// Package opc reads and writes Open Packaging Convention (OPC)
// packages — the ZIP container layer of .docx files.
//
// Open enforces the OPC-07 safety limits before any payload is
// decompressed; part payloads are read lazily through Part.Open with
// size caps. Save emits canonically ordered, deterministic packages and
// passes untouched parts through byte-identical.
package opc

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Conformance records the OOXML conformance class detected on open
// (OPC-05). Save writes Transitional for new packages and preserves
// the source class when editing.
type Conformance int

const (
	// Transitional is the classic schemas.openxmlformats.org family.
	Transitional Conformance = iota
	// Strict is the purl.oclc.org URI family.
	Strict
)

func (c Conformance) String() string {
	if c == Strict {
		return "Strict"
	}
	return "Transitional"
}

// Part is a single package entry. Payload access is lazy and
// size-capped via Open.
type Part struct {
	Name string

	file     *zip.File // non-nil when sourced from the original archive
	data     []byte    // replacement payload when modified
	modified bool
	deleted  bool
}

// Open returns a ReadCloser over the part payload, capped at
// MaxPartBytes. Reads beyond the cap yield ErrDecompressionLimit.
func (p *Part) Open() (io.ReadCloser, error) {
	if p.deleted {
		return nil, fmt.Errorf("opc: part %s deleted: %w", p.Name, ErrInvalidPackage)
	}
	if p.modified {
		return io.NopCloser(&capReader{r: strings.NewReader(string(p.data))}), nil
	}
	rc, err := p.file.Open()
	if err != nil {
		return nil, fmt.Errorf("opc: open part %s: %w", p.Name, err)
	}
	return &cappedReadCloser{rc: rc, cr: &capReader{r: rc}}, nil
}

// capReader returns ErrDecompressionLimit once MaxPartBytes is exceeded.
type capReader struct {
	r   io.Reader
	n   int64
	err error
}

func (c *capReader) Read(p []byte) (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	n, err := c.r.Read(p)
	c.n += int64(n)
	if c.n > MaxPartBytes {
		c.err = fmt.Errorf("opc: part exceeds %d bytes: %w",
			int64(MaxPartBytes), ErrDecompressionLimit)
		return n, c.err
	}
	return n, err
}

type cappedReadCloser struct {
	rc io.Closer
	cr *capReader
}

func (c *cappedReadCloser) Read(p []byte) (int, error) { return c.cr.Read(p) }
func (c *cappedReadCloser) Close() error               { return c.rc.Close() }

// Package is an opened OPC package.
type Package struct {
	zipReader   *zip.Reader
	Parts       map[string]*Part
	ContentTypes *ContentTypes
	// Rels maps source part name ("" for the package root) to its
	// relationship set.
	Rels        map[string]*Relationships
	Conformance Conformance
	warnings    []string
}

// Warnings returns non-fatal issues found on open (D-12) — e.g.
// external relationships that are preserved but never fetched.
func (p *Package) Warnings() []string { return p.warnings }

// MarkModified marks a part's payload as replaced; Save will serialize
// data instead of raw-copying the original entry.
func (p *Package) MarkModified(name string, data []byte) {
	if part, ok := p.Parts[name]; ok {
		part.data = data
		part.modified = true
	}
}

// MarkDeleted removes a part from the saved output.
func (p *Package) MarkDeleted(name string) {
	if part, ok := p.Parts[name]; ok {
		part.deleted = true
	}
}

// Open reads an OPC package from r, enforcing all OPC-07 safety limits
// before parsing: entry-count cap, part-name validation, decompressed
// size caps, and compression-ratio flagging. [Content_Types].xml and
// all .rels parts are parsed eagerly; part payloads stay lazy.
func Open(r io.ReaderAt, size int64) (*Package, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("opc: open zip: %w", err)
	}

	if len(zr.File) > MaxParts {
		return nil, fmt.Errorf("opc: %d entries exceeds cap %d: %w",
			len(zr.File), MaxParts, ErrTooManyParts)
	}

	pkg := &Package{
		zipReader: zr,
		Parts:     make(map[string]*Part, len(zr.File)),
		Rels:      make(map[string]*Relationships),
	}

	var totalDeclared int64
	for _, f := range zr.File {
		if err := validatePartName(f.Name); err != nil {
			return nil, err
		}
		if f.UncompressedSize64 > MaxPartBytes {
			return nil, fmt.Errorf("opc: part %s declares %d bytes (cap %d): %w",
				f.Name, f.UncompressedSize64, int64(MaxPartBytes), ErrDecompressionLimit)
		}
		totalDeclared += int64(f.UncompressedSize64)
		if totalDeclared > MaxTotalBytes {
			return nil, fmt.Errorf("opc: package declares %d bytes (cap %d): %w",
				totalDeclared, int64(MaxTotalBytes), ErrDecompressionLimit)
		}
		if f.CompressedSize64 > 0 && f.CompressedSize64 < 1024 &&
			f.UncompressedSize64/f.CompressedSize64 > MaxCompressionRatio {
			pkg.warnings = append(pkg.warnings, fmt.Sprintf(
				"part %s: compression ratio %d:1 exceeds %d:1",
				f.Name, f.UncompressedSize64/f.CompressedSize64, MaxCompressionRatio))
		}
		pkg.Parts[f.Name] = &Part{Name: f.Name, file: f}
	}

	// [Content_Types].xml is mandatory.
	ctPart, ok := pkg.Parts["[Content_Types].xml"]
	if !ok {
		return nil, fmt.Errorf("opc: missing [Content_Types].xml: %w", ErrInvalidPackage)
	}
	ctBytes, err := readAllCapped(ctPart)
	if err != nil {
		return nil, err
	}
	pkg.ContentTypes, err = parseContentTypes(strings.NewReader(string(ctBytes)))
	if err != nil {
		return nil, err
	}

	// Parse all .rels parts eagerly.
	for name, part := range pkg.Parts {
		if !isRelsPath(name) {
			continue
		}
		b, err := readAllCapped(part)
		if err != nil {
			return nil, err
		}
		rs, err := parseRelationships(strings.NewReader(string(b)))
		if err != nil {
			return nil, fmt.Errorf("opc: %s: %w", name, err)
		}
		src := sourcePartFor(name)
		pkg.Rels[src] = rs
		for _, rel := range rs.Rels {
			if rel.TargetMode == "External" {
				pkg.warnings = append(pkg.warnings, fmt.Sprintf(
					"external relationship %s -> %s (never fetched)", rel.ID, rel.Target))
			}
		}
	}
	if _, ok := pkg.Rels[""]; !ok {
		return nil, fmt.Errorf("opc: missing _rels/.rels: %w", ErrInvalidPackage)
	}

	pkg.Conformance = pkg.detectConformance()
	return pkg, nil
}

// readAllCapped reads a part payload fully, honoring the per-part cap.
func readAllCapped(p *Part) ([]byte, error) {
	rc, err := p.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("opc: read part %s: %w", p.Name, err)
	}
	return b, nil
}

// detectConformance peeks at word/document.xml for the Strict URI
// family. Absence of the part (or of Strict URIs) means Transitional.
func (p *Package) detectConformance() Conformance {
	part, ok := p.Parts["word/document.xml"]
	if !ok {
		return Transitional
	}
	rc, err := part.Open()
	if err != nil {
		return Transitional
	}
	defer rc.Close()
	buf := make([]byte, 64<<10)
	n, _ := io.ReadFull(rc, buf)
	if strings.Contains(string(buf[:n]), "http://purl.oclc.org/ooxml/") {
		return Strict
	}
	return Transitional
}

// validatePartName rejects names that escape the package or are
// malformed (OPC-07, T-01-02). Policy mirrors archive/zip's
// ErrInsecurePath plus OPC part-name grammar.
func validatePartName(name string) error {
	bad := func(why string) error {
		return fmt.Errorf("opc: part name %q: %s: %w", name, why, ErrUnsafePath)
	}
	if name == "" {
		return bad("empty name")
	}
	if !utf8.ValidString(name) {
		return bad("not valid UTF-8")
	}
	if strings.HasPrefix(name, "/") {
		return bad("absolute path")
	}
	if strings.Contains(name, "\\") {
		return bad("backslash separator")
	}
	if len(name) >= 2 && name[1] == ':' {
		return bad("drive letter")
	}
	for _, seg := range strings.Split(name, "/") {
		if seg == "" {
			return bad("empty path segment")
		}
		if seg == ".." {
			return bad("parent traversal")
		}
	}
	return nil
}
