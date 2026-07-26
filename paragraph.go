package wordingo

import (
	"strings"

	"github.com/fabiomarini/wordingo/internal/wml"
)

// Paragraph wraps a WordprocessingML paragraph (w:p).
type Paragraph struct {
	ct  *wml.CT_P
	doc *Document
}

// Style returns the paragraph style ID, or "" if none.
func (p *Paragraph) Style() string {
	if p.ct.PPr != nil && p.ct.PPr.PStyle != nil && p.ct.PPr.PStyle.Val != nil {
		return *p.ct.PPr.PStyle.Val
	}
	return ""
}

// Text returns all text content concatenated across runs.
func (p *Paragraph) Text() string {
	var b strings.Builder
	for _, r := range p.ct.R {
		if r.T != nil {
			b.WriteString(r.T.Value)
		}
	}
	return b.String()
}

// X returns the underlying CT_P for escape-hatch access.
func (p *Paragraph) X() *wml.CT_P {
	if p == nil {
		panic("wordingo: X called on nil Paragraph")
	}
	return p.ct
}

// AddRun appends a run with text to the paragraph and returns it.
func (p *Paragraph) AddRun(text string) *Run {
	if p == nil {
		panic("wordingo: AddRun called on nil Paragraph")
	}
	t := &wml.CT_Text{Value: text}
	ct := &wml.CT_R{T: t}
	p.ct.R = append(p.ct.R, ct)
	p.doc.dirty = true
	return &Run{ct: ct, para: p}
}

// SetAlignment sets paragraph alignment.
func (p *Paragraph) SetAlignment(a Alignment) *Paragraph {
	if p == nil {
		panic("wordingo: SetAlignment called on nil Paragraph")
	}
	if p.ct.PPr == nil {
		p.ct.PPr = &wml.CT_PPr{}
	}
	val := a.String()
	if val == "" {
		p.doc.warn("wordingo: unknown alignment %d", a)
		return p
	}
	p.ct.PPr.Jc = &wml.CT_Jc{Val: &val}
	p.doc.dirty = true
	return p
}

// SetSpacing sets paragraph spacing.
func (p *Paragraph) SetSpacing(s *ParSpacing) *Paragraph {
	if p == nil {
		panic("wordingo: SetSpacing called on nil Paragraph")
	}
	if s == nil {
		p.doc.warn("wordingo: ParSpacing is nil")
		return p
	}
	if p.ct.PPr == nil {
		p.ct.PPr = &wml.CT_PPr{}
	}
	sp := &wml.CT_Spacing{}
	if s.Before != 0 {
		sp.Before = &s.Before
	}
	if s.After != 0 {
		sp.After = &s.After
	}
	if s.Line != 0 {
		sp.Line = &s.Line
	}
	if s.LineRule != "" {
		sp.LineRule = &s.LineRule
	}
	p.ct.PPr.Spacing = sp
	p.doc.dirty = true
	return p
}

// SetIndent sets paragraph indentation.
func (p *Paragraph) SetIndent(i *ParIndent) *Paragraph {
	if p == nil {
		panic("wordingo: SetIndent called on nil Paragraph")
	}
	if i == nil {
		p.doc.warn("wordingo: ParIndent is nil")
		return p
	}
	if p.ct.PPr == nil {
		p.ct.PPr = &wml.CT_PPr{}
	}
	ind := &wml.CT_Ind{}
	if i.Left != 0 {
		ind.Left = &i.Left
	}
	if i.Right != 0 {
		ind.Right = &i.Right
	}
	if i.FirstLine != 0 {
		ind.FirstLine = &i.FirstLine
	}
	if i.Hanging != 0 {
		ind.Hanging = &i.Hanging
	}
	p.ct.PPr.Ind = ind
	p.doc.dirty = true
	return p
}

// SetStyle sets the paragraph style reference.
// Empty string clears the style reference.
func (p *Paragraph) SetStyle(name string) *Paragraph {
	if p == nil {
		panic("wordingo: SetStyle called on nil Paragraph")
	}
	if name == "" {
		if p.ct.PPr != nil {
			p.ct.PPr.PStyle = nil
		}
		p.doc.dirty = true
		return p
	}
	if p.ct.PPr == nil {
		p.ct.PPr = &wml.CT_PPr{}
	}
	p.ct.PPr.PStyle = &wml.CT_PStyle{Val: &name}
	p.doc.dirty = true
	return p
}

// SetPageBreakBefore sets or clears the page-break-before property
// on the paragraph.  When enabled, the paragraph always starts on a
// new page.
func (p *Paragraph) SetPageBreakBefore(b bool) *Paragraph {
	if p == nil {
		panic("wordingo: SetPageBreakBefore called on nil Paragraph")
	}
	if p.ct.PPr == nil {
		p.ct.PPr = &wml.CT_PPr{}
	}
	if b {
		p.ct.PPr.PageBreakBefore = &wml.CT_OnOff{}
	} else {
		p.ct.PPr.PageBreakBefore = nil
	}
	p.doc.dirty = true
	return p
}

// SetFormatting sets multiple paragraph formatting properties.
func (p *Paragraph) SetFormatting(f ParFormat) *Paragraph {
	if p == nil {
		panic("wordingo: SetFormatting called on nil Paragraph")
	}
	if f.Alignment != nil {
		p.SetAlignment(*f.Alignment)
	}
	if f.Spacing != nil {
		p.SetSpacing(f.Spacing)
	}
	if f.Indent != nil {
		p.SetIndent(f.Indent)
	}
	return p
}
