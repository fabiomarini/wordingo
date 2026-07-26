package wordingo

import (
	"github.com/fabiomarini/wordingo/internal/wml"
)

// Section wraps a section properties block (w:sectPr).  In v1 the
// document has a single section; the Section type provides a forward-
// compatible API for multi-section support in a future version.
type Section struct {
	doc *Document
	ct  *wml.CT_SectPr
}

// X returns the underlying CT_SectPr for escape-hatch access.
func (s *Section) X() *wml.CT_SectPr {
	if s == nil {
		panic("wordingo: X called on nil Section")
	}
	return s.ct
}

// SetOrientation sets the page orientation.  Landscape swaps W and H
// on the current PgSz; Portrait restores the W < H default.  The
// underlying OOXML stores orientation via the W/H ratio rather than a
// separate attribute (T-05-07 — accepted behavior).
func (s *Section) SetOrientation(o PageOrientation) {
	if s == nil {
		panic("wordingo: SetOrientation called on nil Section")
	}
	if s.ct.PgSz == nil {
		s.ct.PgSz = &wml.CT_PgSz{W: ptrInt64(PaperLetterW), H: ptrInt64(PaperLetterH)}
	}
	w, h := *s.ct.PgSz.W, *s.ct.PgSz.H
	switch o {
	case OrientationLandscape:
		if w < h {
			// Swap so W > H for landscape.
			*s.ct.PgSz.W = h
			*s.ct.PgSz.H = w
		}
	default: // Portrait
		if w > h {
			// Swap so W < H for portrait.
			*s.ct.PgSz.W = h
			*s.ct.PgSz.H = w
		}
	}
	s.doc.dirty = true
}

// SetPaperSize sets the page dimensions in twips.  Convenience constants
// PaperLetterW/H, PaperA4W/H, and PaperLegalW/H are available.
func (s *Section) SetPaperSize(w, h int64) {
	if s == nil {
		panic("wordingo: SetPaperSize called on nil Section")
	}
	if s.ct.PgSz == nil {
		s.ct.PgSz = &wml.CT_PgSz{}
	}
	s.ct.PgSz.W = &w
	s.ct.PgSz.H = &h
	s.doc.dirty = true
}

// SetMargins sets the page margins in twips (top, right, bottom, left).
// (1 inch = 1440 twips; 1 cm ≈ 567 twips).
func (s *Section) SetMargins(top, right, bottom, left int64) {
	if s == nil {
		panic("wordingo: SetMargins called on nil Section")
	}
	if s.ct.PgMar == nil {
		s.ct.PgMar = &wml.CT_PgMar{}
	}
	s.ct.PgMar.Top = &top
	s.ct.PgMar.Right = &right
	s.ct.PgMar.Bottom = &bottom
	s.ct.PgMar.Left = &left
	s.doc.dirty = true
}

// ---- Document page setup methods ----

// Section returns the Section wrapper over the document's section
// properties.  The section is lazily initialised if nil.
func (d *Document) Section() *Section {
	if d == nil {
		panic("wordingo: Section called on nil Document")
	}
	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	if d.doc.Body.SectPr == nil {
		d.doc.Body.SectPr = defaultSectPr()
	}
	return &Section{doc: d, ct: d.doc.Body.SectPr}
}

// SetOrientation sets the page orientation on the document's section
// and returns the Document for method chaining.
func (d *Document) SetOrientation(o PageOrientation) *Document {
	d.Section().SetOrientation(o)
	return d
}

// SetPaperSize sets the page dimensions on the document's section
// and returns the Document for method chaining.
func (d *Document) SetPaperSize(w, h int64) *Document {
	d.Section().SetPaperSize(w, h)
	return d
}

// SetMargins sets the page margins on the document's section
// (top, right, bottom, left twips) and returns the Document for
// method chaining.
func (d *Document) SetMargins(top, right, bottom, left int64) *Document {
	d.Section().SetMargins(top, right, bottom, left)
	return d
}

// AddPageBreak creates a new empty paragraph with PageBreakBefore set
// and appends it to the body.  This forces the following content onto
// a new page.  Returns the paragraph for further formatting.
func (d *Document) AddPageBreak() *Paragraph {
	if d == nil {
		panic("wordingo: AddPageBreak called on nil Document")
	}
	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}
	ct := &wml.CT_P{
		PPr: &wml.CT_PPr{
			PageBreakBefore: &wml.CT_OnOff{},
		},
	}
	d.doc.Body.P = append(d.doc.Body.P, ct)
	d.dirty = true
	return &Paragraph{ct: ct, doc: d}
}
