package wordingo

import (
	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
)

// AddHyperlink creates a hyperlink on the paragraph with the given display
// text and target URI. A new relationship is created in
// word/_rels/document.xml.rels with TargetMode="External" (Pitfall 7).
// Each call allocates a fresh rId via NextRID() — duplicate URIs are not
// deduplicated (D-23 agent discretion; each extra rId ~50 bytes).
//
// Returns a *Run that supports chaining formatting methods
// (SetBold, SetColor, etc.).
//
//	para.AddHyperlink("click here", "https://example.com").SetBold(true).SetColor("0563C1")
func (p *Paragraph) AddHyperlink(text, uri string) *Run {
	if p == nil {
		panic("wordingo: AddHyperlink called on nil Paragraph")
	}

	rels := p.doc.pkg.Rels["word/document.xml"]
	if rels == nil {
		rels = &opc.Relationships{}
		p.doc.pkg.Rels["word/document.xml"] = rels
	}

	rID := rels.NextRID()
	rels.Rels = append(rels.Rels, opc.Relationship{
		ID:         rID,
		Type:       relHyperlink,
		Target:     uri,
		TargetMode: "External",
	})

	run := &wml.CT_R{T: &wml.CT_Text{Value: text}}
	hl := &wml.CT_Hyperlink{
		ID: rID,
		R:  []*wml.CT_R{run},
	}
	p.ct.Hyperlink = append(p.ct.Hyperlink, hl)
	p.doc.dirty = true

	return &Run{ct: run, para: p}
}
