package wordingo

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

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
	d := &Document{
		pkg:          dst,
		doc:          doc,
		nextImageID:  1,
		nextHeaderID: 1,
		nextFooterID: 1,
	}
	d.syncBodyOrder()
	return d, nil
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

	srcRels := src.Rels["word/document.xml"]
	dstRels := dst.Rels["word/document.xml"]
	if dstRels == nil {
		dstRels = &opc.Relationships{}
		dst.Rels["word/document.xml"] = dstRels
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

	// Clone header/footer parts referenced in source sectPr (D-13,
	// reverse Phase 3 Pitfall 4).  Allocate fresh rIds via dstRels
	// to avoid collisions (T-05-05).
	srcSectPr := srcDoc.Body.SectPr
	sectPr := defaultSectPr()

	if srcSectPr != nil {
		// Clone header references.
		for _, ref := range srcSectPr.HdrFtrRef {
			if srcRels == nil {
				continue
			}
			srcRel := findRelByID(srcRels, ref.ID)
			if srcRel == nil {
				continue
			}
			target := path.Join("word", srcRel.Target)
			srcPart, ok := src.Parts[target]
			if !ok {
				continue
			}
			partBytes, err := readPartBytes(srcPart)
			if err != nil {
				continue
			}

			// Copy header part to dst with fresh rId.
			dst.MarkModified(target, partBytes)
			dst.ContentTypes.Overrides["/"+target] = ctHeader
			newID := dstRels.NextRID()
			dstRels.Rels = append(dstRels.Rels, opc.Relationship{
				ID:     newID,
				Type:   relHeader,
				Target: srcRel.Target,
			})

			sectPr.HdrFtrRef = append(sectPr.HdrFtrRef, &wml.CT_HdrFtrRef{
				ID:   newID,
				Type: ref.Type,
			})
		}

		// Clone footer references.
		for _, ref := range srcSectPr.FtrRef {
			if srcRels == nil {
				continue
			}
			srcRel := findRelByID(srcRels, ref.ID)
			if srcRel == nil {
				continue
			}
			target := path.Join("word", srcRel.Target)
			srcPart, ok := src.Parts[target]
			if !ok {
				continue
			}
			partBytes, err := readPartBytes(srcPart)
			if err != nil {
				continue
			}

			dst.MarkModified(target, partBytes)
			dst.ContentTypes.Overrides["/"+target] = ctFooter
			newID := dstRels.NextRID()
			dstRels.Rels = append(dstRels.Rels, opc.Relationship{
				ID:     newID,
				Type:   relFooter,
				Target: srcRel.Target,
			})

			sectPr.FtrRef = append(sectPr.FtrRef, &wml.CT_HdrFtrRef{
				ID:   newID,
				Type: ref.Type,
			})
		}

		// Preserve TitlePg if source has it.
		sectPr.TitlePg = srcSectPr.TitlePg
	}

	// Re-use source PgSz/PgMar (if present) instead of defaults.
	if srcSectPr != nil {
		if srcSectPr.PgSz != nil {
			sectPr.PgSz = srcSectPr.PgSz
		}
		if srcSectPr.PgMar != nil {
			sectPr.PgMar = srcSectPr.PgMar
		}
	}

	freshDoc := &wml.CT_Document{
		Body: &wml.CT_Body{
			P:      srcDoc.Body.P,
			Tbl:    srcDoc.Body.Tbl,
			SectPr: sectPr,
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
	d := &Document{
		pkg:          dst,
		doc:          doc,
		nextImageID:  1,
		nextHeaderID: 1,
		nextFooterID: 1,
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
