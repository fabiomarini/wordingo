package wordingo

import (
	"strings"

	"github.com/fabiomarini/wordingo/internal/wml"
)

// Paragraph wraps a WordprocessingML paragraph (w:p).
// Accessors are read-only in Phase 3; Phase 4 adds mutation.
type Paragraph struct {
	ct *wml.CT_P
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
func (p *Paragraph) X() *wml.CT_P { return p.ct }
