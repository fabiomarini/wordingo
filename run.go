package wordingo

import (
	"math"
	"strings"

	"github.com/fabiomarini/wordingo/internal/wml"
)

type Run struct {
	ct   *wml.CT_R
	para *Paragraph
}

func (r *Run) X() *wml.CT_R {
	if r == nil {
		panic("wordingo: X called on nil Run")
	}
	return r.ct
}

func (r *Run) SetBold(b bool) *Run {
	if r == nil {
		panic("wordingo: SetBold called on nil Run")
	}
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.B = &wml.CT_OnOff{Val: &b}
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetItalic(b bool) *Run {
	if r == nil {
		panic("wordingo: SetItalic called on nil Run")
	}
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.I = &wml.CT_OnOff{Val: &b}
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetUnderline(u string) *Run {
	if r == nil {
		panic("wordingo: SetUnderline called on nil Run")
	}
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.U = &wml.CT_U{Val: &u}
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetFont(name string) *Run {
	if r == nil {
		panic("wordingo: SetFont called on nil Run")
	}
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.RFonts = &wml.CT_RFonts{Ascii: &name, HAnsi: &name}
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetSize(pts float64) *Run {
	if r == nil {
		panic("wordingo: SetSize called on nil Run")
	}
	if pts < 0 {
		r.para.doc.warn("wordingo: negative font size %f", pts)
		return r
	}
	halfPts := int64(math.Round(pts * 2))
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.Sz = &wml.CT_Sz{Val: &halfPts}
	r.para.doc.dirty = true
	return r
}

func validHexColor(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9') && !(c >= 'A' && c <= 'F') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func (r *Run) SetColor(hex string) *Run {
	if r == nil {
		panic("wordingo: SetColor called on nil Run")
	}
	if !validHexColor(hex) {
		r.para.doc.warn("wordingo: invalid color %q: must be 6 hex digits", hex)
		return r
	}
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.Color = &wml.CT_Color{Val: &hex}
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetHighlight(color string) *Run {
	if r == nil {
		panic("wordingo: SetHighlight called on nil Run")
	}
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.Highlight = &wml.CT_Highlight{Val: &color}
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetStyle(name string) *Run {
	if r == nil {
		panic("wordingo: SetStyle called on nil Run")
	}
	if name == "" {
		if r.ct.RPr != nil {
			r.ct.RPr.RStyle = nil
		}
		r.para.doc.dirty = true
		return r
	}
	if r.ct.RPr == nil {
		r.ct.RPr = &wml.CT_RPr{}
	}
	r.ct.RPr.RStyle = &wml.CT_RStyle{Val: &name}
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetText(s string) *Run {
	if r == nil {
		panic("wordingo: SetText called on nil Run")
	}
	if r.ct.T == nil {
		r.ct.T = &wml.CT_Text{Value: s}
	} else {
		r.ct.T.Value = s
	}
	r.para.doc.dirty = true
	return r
}

func (r *Run) ReplaceText(old, new string) *Run {
	if r == nil {
		panic("wordingo: ReplaceText called on nil Run")
	}
	if r.ct.T == nil {
		return r
	}
	r.ct.T.Value = strings.ReplaceAll(r.ct.T.Value, old, new)
	r.para.doc.dirty = true
	return r
}

func (r *Run) SetFormatting(f RunFormat) *Run {
	if r == nil {
		panic("wordingo: SetFormatting called on nil Run")
	}
	if f.Bold != nil {
		r.SetBold(*f.Bold)
	}
	if f.Italic != nil {
		r.SetItalic(*f.Italic)
	}
	if f.Underline != nil {
		r.SetUnderline(*f.Underline)
	}
	if f.Font != nil {
		r.SetFont(*f.Font)
	}
	if f.Size != nil {
		r.SetSize(*f.Size)
	}
	if f.Color != nil {
		r.SetColor(*f.Color)
	}
	if f.Highlight != nil {
		r.SetHighlight(*f.Highlight)
	}
	return r
}
