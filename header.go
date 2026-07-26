package wordingo

import (
	"bytes"
	"fmt"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// ParagraphContainer is the interface for editing paragraphs in a
// container that owns a paragraph slice (body, header, or footer).
type ParagraphContainer interface {
	Paragraphs() []*Paragraph
	InsertParagraphAt(idx int, ct *wml.CT_P) *Paragraph
	DeleteParagraphAt(idx int)
}

// Header wraps a WordprocessingML header part (w:hdr).
type Header struct {
	ct       *wml.CT_Hdr
	doc      *Document
	partName string
}

// Footer wraps a WordprocessingML footer part (w:ftr).
type Footer struct {
	ct       *wml.CT_Ftr
	doc      *Document
	partName string
}

// X returns the underlying CT_Hdr for escape-hatch access.
func (h *Header) X() *wml.CT_Hdr {
	if h == nil {
		panic("wordingo: X called on nil Header")
	}
	return h.ct
}

// X returns the underlying CT_Ftr for escape-hatch access.
func (f *Footer) X() *wml.CT_Ftr {
	if f == nil {
		panic("wordingo: X called on nil Footer")
	}
	return f.ct
}

func (h *Header) Paragraphs() []*Paragraph {
	if h == nil {
		panic("wordingo: Paragraphs called on nil Header")
	}
	paras := make([]*Paragraph, len(h.ct.P))
	for i, p := range h.ct.P {
		paras[i] = &Paragraph{ct: p, doc: h.doc}
	}
	return paras
}

func (h *Header) InsertParagraphAt(idx int, ct *wml.CT_P) *Paragraph {
	if h == nil {
		panic("wordingo: InsertParagraphAt called on nil Header")
	}
	h.ct.P = append(h.ct.P, nil)
	copy(h.ct.P[idx+1:], h.ct.P[idx:])
	h.ct.P[idx] = ct
	h.sync()
	return &Paragraph{ct: ct, doc: h.doc}
}

func (h *Header) DeleteParagraphAt(idx int) {
	if h == nil {
		panic("wordingo: DeleteParagraphAt called on nil Header")
	}
	h.ct.P = append(h.ct.P[:idx], h.ct.P[idx+1:]...)
	h.sync()
}

func (h *Header) sync() {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(h.ct); err != nil {
		h.doc.warn("wordingo: encode header: %v", err)
		return
	}
	if err := enc.Flush(); err != nil {
		h.doc.warn("wordingo: flush header: %v", err)
		return
	}
	h.doc.pkg.MarkModified(h.partName, buf.Bytes())
	h.doc.dirty = true
}

// AddParagraph appends a paragraph with optional text to the header
// and returns it.  The paragraph is added as a child of the header
// element (w:hdr/w:p).
func (h *Header) AddParagraph(text string) *Paragraph {
	if h == nil {
		panic("wordingo: AddParagraph called on nil Header")
	}
	ct := &wml.CT_P{}
	if text != "" {
		ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
	}
	h.ct.P = append(h.ct.P, ct)
	h.sync()
	return &Paragraph{ct: ct, doc: h.doc}
}

func (f *Footer) Paragraphs() []*Paragraph {
	if f == nil {
		panic("wordingo: Paragraphs called on nil Footer")
	}
	paras := make([]*Paragraph, len(f.ct.P))
	for i, p := range f.ct.P {
		paras[i] = &Paragraph{ct: p, doc: f.doc}
	}
	return paras
}

func (f *Footer) InsertParagraphAt(idx int, ct *wml.CT_P) *Paragraph {
	if f == nil {
		panic("wordingo: InsertParagraphAt called on nil Footer")
	}
	f.ct.P = append(f.ct.P, nil)
	copy(f.ct.P[idx+1:], f.ct.P[idx:])
	f.ct.P[idx] = ct
	f.sync()
	return &Paragraph{ct: ct, doc: f.doc}
}

func (f *Footer) DeleteParagraphAt(idx int) {
	if f == nil {
		panic("wordingo: DeleteParagraphAt called on nil Footer")
	}
	f.ct.P = append(f.ct.P[:idx], f.ct.P[idx+1:]...)
	f.sync()
}

func (f *Footer) sync() {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(f.ct); err != nil {
		f.doc.warn("wordingo: encode footer: %v", err)
		return
	}
	if err := enc.Flush(); err != nil {
		f.doc.warn("wordingo: flush footer: %v", err)
		return
	}
	f.doc.pkg.MarkModified(f.partName, buf.Bytes())
	f.doc.dirty = true
}

// AddParagraph appends a paragraph with optional text to the footer
// and returns it.  The paragraph is added as a child of the footer
// element (w:ftr/w:p).
func (f *Footer) AddParagraph(text string) *Paragraph {
	if f == nil {
		panic("wordingo: AddParagraph called on nil Footer")
	}
	ct := &wml.CT_P{}
	if text != "" {
		ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: text}}}
	}
	f.ct.P = append(f.ct.P, ct)
	f.sync()
	return &Paragraph{ct: ct, doc: f.doc}
}

// addHelperPart creates a new OPC part in the document package and adds
// a relationship entry and content type override.  Returns the new rId.
// The part name is the full internal path (e.g. "word/header1.xml"),
// relType is the OPC relationship type (e.g. relHeader), ctType is the
// MIME content type (e.g. ctHeader), and targetSuffix is the relative
// target stored in the relationship (e.g. "header1.xml").
func (d *Document) addHelperPart(partName string, data []byte, relType, ctType, targetSuffix string) string {
	if d == nil {
		panic("wordingo: addHelperPart called on nil Document")
	}

	d.pkg.MarkModified(partName, data)
	d.pkg.ContentTypes.Overrides["/"+partName] = ctType

	rels := d.pkg.Rels["word/document.xml"]
	if rels == nil {
		rels = &opc.Relationships{}
		d.pkg.Rels["word/document.xml"] = rels
	}
	rID := rels.NextRID()
	rels.Rels = append(rels.Rels, opc.Relationship{
		ID:     rID,
		Type:   relType,
		Target: targetSuffix,
	})

	d.dirty = true
	return rID
}

// AddHeader creates a header part with the given variant (default, first,
// or even), registers it in the OPC package, links it to the section
// properties via a headerReference element, and returns the Header.
func (d *Document) AddHeader(variant HeaderVariant) *Header {
	if d == nil {
		panic("wordingo: AddHeader called on nil Document")
	}

	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	if d.doc.Body.SectPr == nil {
		d.doc.Body.SectPr = defaultSectPr()
	}

	// Build CT_Hdr with one empty Normal paragraph.
	valNormal := "Normal"
	hdr := &wml.CT_Hdr{
		P: []*wml.CT_P{
			{
				PPr: &wml.CT_PPr{
					PStyle: &wml.CT_PStyle{Val: &valNormal},
				},
			},
		},
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(hdr); err != nil {
		d.warn("wordingo: encode header: %v", err)
		return nil
	}
	if err := enc.Flush(); err != nil {
		d.warn("wordingo: flush header: %v", err)
		return nil
	}

	partName := fmt.Sprintf("word/header%d.xml", d.nextHeaderID)
	targetSuffix := fmt.Sprintf("header%d.xml", d.nextHeaderID)
	d.nextHeaderID++

	rID := d.addHelperPart(partName, buf.Bytes(), relHeader, ctHeader, targetSuffix)

	// Link to section properties.
	ref := &wml.CT_HdrFtrRef{
		ID:   rID,
		Type: variant.String(),
	}
	d.doc.Body.SectPr.HdrFtrRef = append(d.doc.Body.SectPr.HdrFtrRef, ref)

	d.dirty = true
	return &Header{ct: hdr, doc: d, partName: partName}
}

// AddFooter creates a footer part with the given variant (default, first,
// or even), registers it in the OPC package, links it to the section
// properties via a footerReference element, and returns the Footer.
func (d *Document) AddFooter(variant FooterVariant) *Footer {
	if d == nil {
		panic("wordingo: AddFooter called on nil Document")
	}

	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	if d.doc.Body.SectPr == nil {
		d.doc.Body.SectPr = defaultSectPr()
	}

	// Build CT_Ftr with one empty Normal paragraph.
	valNormal := "Normal"
	ftr := &wml.CT_Ftr{
		P: []*wml.CT_P{
			{
				PPr: &wml.CT_PPr{
					PStyle: &wml.CT_PStyle{Val: &valNormal},
				},
			},
		},
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(ftr); err != nil {
		d.warn("wordingo: encode footer: %v", err)
		return nil
	}
	if err := enc.Flush(); err != nil {
		d.warn("wordingo: flush footer: %v", err)
		return nil
	}

	partName := fmt.Sprintf("word/footer%d.xml", d.nextFooterID)
	targetSuffix := fmt.Sprintf("footer%d.xml", d.nextFooterID)
	d.nextFooterID++

	rID := d.addHelperPart(partName, buf.Bytes(), relFooter, ctFooter, targetSuffix)

	// Link to section properties.
	ref := &wml.CT_HdrFtrRef{
		ID:   rID,
		Type: variant.String(),
	}
	d.doc.Body.SectPr.FtrRef = append(d.doc.Body.SectPr.FtrRef, ref)

	d.dirty = true
	return &Footer{ct: ftr, doc: d, partName: partName}
}
