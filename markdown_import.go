package wordingo

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/fabiomarini/wordingo/internal/markdown"
	"github.com/fabiomarini/wordingo/internal/wml"
)

func CreateFromMarkdown(input string) (*Document, error) {
	d, err := Create()
	if err != nil {
		return nil, err
	}

	blocks, err := markdown.Parse(input)
	if err != nil {
		return nil, err
	}

	importBlocks(d, blocks)
	return d, nil
}

func (d *Document) ImportMarkdown(input string) {
	if d == nil {
		panic("wordingo: ImportMarkdown called on nil Document")
	}

	blocks, err := markdown.Parse(input)
	if err != nil {
		d.warn("wordingo: markdown parse error: %v", err)
		return
	}

	importBlocks(d, blocks)
}

func importBlocks(d *Document, blocks []markdown.Block) {
	if d.doc.Body == nil {
		d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
	}

	for _, block := range blocks {
		switch block.Type {
		case markdown.BlockParagraph:
			importParagraph(d, block)
		case markdown.BlockHeading:
			importHeading(d, block)
		case markdown.BlockCodeBlock:
			importCodeBlock(d, block)
		case markdown.BlockOrderedList:
			importList(d, block, true)
		case markdown.BlockBulletList:
			importList(d, block, false)
		case markdown.BlockTable:
			importTable(d, block)
		}
	}

	d.dirty = true
}

func importParagraph(d *Document, block markdown.Block) {
	p := d.AddParagraph("")
	if len(block.Inlines) > 0 {
		applyInlines(p, block.Inlines, d)
	} else {
		p.AddRun(block.Content)
	}
}

func importHeading(d *Document, block markdown.Block) {
	styleName := fmt.Sprintf("Heading%d", block.Level)
	p := d.AddParagraph("").SetStyle(styleName)
	if len(block.Inlines) > 0 {
		applyInlines(p, block.Inlines, d)
	} else {
		p.AddRun(block.Content)
	}
}

func importCodeBlock(d *Document, block markdown.Block) {
	for _, line := range block.Lines {
		p := d.AddParagraph(line)
		for _, r := range p.ct.R {
			if r.RPr == nil {
				r.RPr = &wml.CT_RPr{}
			}
			r.RPr.RFonts = &wml.CT_RFonts{}
			font := "Consolas"
			r.RPr.RFonts.Ascii = &font
			r.RPr.RFonts.HAnsi = &font
		}
	}
}

func importList(d *Document, block markdown.Block, ordered bool) {
	numMap := makeListNumMap(d, ordered)

	for _, item := range block.ListItems {
		ilvl := int64(item.Level)
		ct := &wml.CT_P{
			PPr: &wml.CT_PPr{
				NumPr: &wml.CT_NumPr{
					ILvl:  &wml.CT_ILvl{Val: &ilvl},
					NumId: &wml.CT_NumId{Val: &numMap.numID},
				},
			},
		}

		if len(item.Inlines) > 0 {
			for _, span := range item.Inlines {
				run := &wml.CT_R{T: &wml.CT_Text{Value: span.Text}}
				if span.Bold || span.Italic || span.Code {
					run.RPr = &wml.CT_RPr{}
					if span.Bold {
						run.RPr.B = &wml.CT_OnOff{Val: &yes}
					}
					if span.Italic {
						run.RPr.I = &wml.CT_OnOff{Val: &yes}
					}
					if span.Code {
						font := "Consolas"
						run.RPr.RFonts = &wml.CT_RFonts{Ascii: &font, HAnsi: &font}
					}
				}
				ct.R = append(ct.R, run)

				if span.LinkURL != "" {
					p := &Paragraph{ct: ct, doc: d}
					p.AddHyperlink(span.LinkText, span.LinkURL)
				}
				if span.ImageURL != "" {
					p := &Paragraph{ct: ct, doc: d}
					importImage(d, p, span.ImageURL)
				}
			}
		} else if item.Content != "" {
			ct.R = []*wml.CT_R{{T: &wml.CT_Text{Value: item.Content}}}
		}

		if d.doc.Body == nil {
			d.doc.Body = &wml.CT_Body{SectPr: defaultSectPr()}
		}
		d.doc.Body.AppendP(ct)
	}

	d.dirty = true
}

type listNumMap struct {
	numID         int64
	abstractNumID int64
}

func makeListNumMap(d *Document, ordered bool) listNumMap {
	nb := d.readOrCreateNumbering()

	nextAbsID := findMaxAbstractNumID(nb)
	nextNumID := findMaxNumID(nb)

	levels := makeLevels(ordered)

	absNum := &wml.CT_AbstractNum{
		AbstractNumID: &nextAbsID,
		Lvl:           levels,
	}
	nb.AbstractNum = append(nb.AbstractNum, absNum)

	num := &wml.CT_Num{
		NumID:         &nextNumID,
		AbstractNumID: &wml.CT_AbstractNumID{Val: &nextAbsID},
	}
	nb.Num = append(nb.Num, num)

	d.writeNumbering(nb)

	return listNumMap{numID: nextNumID, abstractNumID: nextAbsID}
}

var yes = true

func importTable(d *Document, block markdown.Block) {
	data := block.Cells
	if len(data) < 2 {
		return
	}

	tb, err := d.AddTable(data)
	if err != nil {
		d.warn("wordingo: skip table: %v", err)
		return
	}

	// AddTable fills cells with plain runs; re-render every cell that has
	// inline formatting (bold/italic/code/links) so table cells match the
	// formatting rules applied to paragraphs and list items.
	ct := tb.X()
	for ri, row := range block.CellInlines {
		if ri >= len(ct.Tr) {
			break
		}
		for ci, spans := range row {
			if ci >= len(ct.Tr[ri].Tc) || len(spans) == 0 {
				continue
			}
			tc := ct.Tr[ri].Tc[ci]
			if len(tc.P) == 0 {
				tc.P = []*wml.CT_P{{}}
			}
			if runs := cellRuns(spans); len(runs) > 0 {
				tc.P[0].R = runs
			}
		}
	}
	d.dirty = true
}

// cellRuns builds formatted runs for one table cell. Hyperlinks fall back
// to their link text (relationship wiring is paragraph-scoped) and images
// are skipped.
func cellRuns(spans []markdown.InlineSpan) []*wml.CT_R {
	var runs []*wml.CT_R
	for _, s := range spans {
		text := s.Text
		if s.LinkURL != "" {
			if s.LinkText != "" {
				text = s.LinkText
			} else {
				text = s.LinkURL
			}
		}
		if s.ImageURL != "" || text == "" {
			continue
		}
		r := &wml.CT_R{T: &wml.CT_Text{Value: text}}
		if s.Bold || s.Italic || s.Code {
			r.RPr = &wml.CT_RPr{}
			if s.Bold {
				r.RPr.B = &wml.CT_OnOff{Val: &yes}
			}
			if s.Italic {
				r.RPr.I = &wml.CT_OnOff{Val: &yes}
			}
			if s.Code {
				font := "Consolas"
				r.RPr.RFonts = &wml.CT_RFonts{Ascii: &font, HAnsi: &font}
			}
		}
		runs = append(runs, r)
	}
	return runs
}

func applyInlines(p *Paragraph, inlines []markdown.InlineSpan, d *Document) {
	for _, span := range inlines {
		if span.LinkURL != "" {
			r := p.AddHyperlink(span.LinkText, span.LinkURL)
			if span.Bold {
				r.SetBold(true)
			}
			if span.Italic {
				r.SetItalic(true)
			}
			continue
		}

		if span.ImageURL != "" {
			importImage(d, p, span.ImageURL)
			continue
		}

		r := p.AddRun(span.Text)
		if span.Bold {
			r.SetBold(true)
		}
		if span.Italic {
			r.SetItalic(true)
		}
		if span.Code {
			r.SetFont("Consolas")
		}
	}
}

func importImage(d *Document, p *Paragraph, dataURI string) {
	decoded, mimeType, err := decodeDataURI(dataURI)
	if err != nil {
		d.warn("wordingo: skip image: %v", err)
		return
	}

	_, err = d.AddImageBytes("image", decoded, mimeType)
	if err != nil {
		d.warn("wordingo: skip image: %v", err)
	}
}

func decodeDataURI(uri string) ([]byte, string, error) {
	idx := strings.Index(uri, ",")
	if idx == -1 {
		return nil, "", fmt.Errorf("wordingo: invalid data URI: no comma")
	}

	meta := uri[:idx]
	b64data := uri[idx+1:]

	var mimeType string
	if strings.HasPrefix(meta, "data:image/png") {
		mimeType = "image/png"
	} else if strings.HasPrefix(meta, "data:image/jpeg") || strings.HasPrefix(meta, "data:image/jpg") {
		mimeType = "image/jpeg"
	} else {
		return nil, "", fmt.Errorf("wordingo: unsupported image type in data URI: %s", meta)
	}

	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return nil, "", fmt.Errorf("wordingo: base64 decode: %w", err)
	}

	return data, mimeType, nil
}
