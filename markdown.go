package wordingo

import (
	"encoding/base64"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/fabiomarini/wordingo/internal/wml"
)

type runFormatKey struct {
	bold   bool
	italic bool
	strike bool
}

func (d *Document) ToMarkdown(opts *ExtractOpts) (string, error) {
	if d == nil {
		panic("wordingo: ToMarkdown called on nil Document")
	}

	if opts == nil {
		opts = &ExtractOpts{ScopedParts: ScopedParts{Body: true, Tables: true, Headers: true, Footers: true}}
	}

	sep := opts.Separator
	if sep == "" {
		sep = "\n"
	}

	var b strings.Builder

	if d.doc.Body == nil {
		return "", nil
	}

	if opts.ScopedParts.Body || opts.ScopedParts.Tables {
		for _, el := range d.Body() {
			switch el.Type {
			case ElementParagraph:
				if opts.ScopedParts.Body {
					md, err := paragraphToMarkdown(el.Para.X(), d)
					if err != nil {
						return "", err
					}
					if md != "" {
						b.WriteString(md)
						b.WriteString(sep)
					}
				}
			case ElementTable:
				if opts.ScopedParts.Tables {
					md := tableToMarkdown(el.Table.X(), d)
					if md != "" {
						b.WriteString(md)
						b.WriteString(sep)
					}
				}
			}
		}
	}

	if opts.ScopedParts.Headers {
		d.readHeaderParts(func(hdr *wml.CT_Hdr) {
			for _, p := range hdr.P {
				md, err := paragraphToMarkdown(p, d)
				if err != nil {
					return
				}
				if md != "" {
					b.WriteString(md)
					b.WriteString(sep)
				}
			}
		})
	}

	if opts.ScopedParts.Footers {
		d.readFooterParts(func(ftr *wml.CT_Ftr) {
			for _, p := range ftr.P {
				md, err := paragraphToMarkdown(p, d)
				if err != nil {
					return
				}
				if md != "" {
					b.WriteString(md)
					b.WriteString(sep)
				}
			}
		})
	}

	result := strings.TrimSuffix(b.String(), sep)
	return result, nil
}

func paragraphToMarkdown(ct_p *wml.CT_P, d *Document) (string, error) {
	if ct_p == nil {
		return "", nil
	}

	var headingLevel int
	if ct_p.PPr != nil && ct_p.PPr.PStyle != nil && ct_p.PPr.PStyle.Val != nil {
		style := *ct_p.PPr.PStyle.Val
		if len(style) > 7 && strings.HasPrefix(style, "Heading") {
			if dig := style[7]; dig >= '1' && dig <= '6' {
				headingLevel = int(dig - '0')
			}
		}
	}

	var prefix string
	if headingLevel > 0 {
		prefix = strings.Repeat("#", headingLevel) + " "
	}

	isCodeBlock := false
	if ct_p.PPr == nil || ct_p.PPr.NumPr == nil {
		for _, r := range ct_p.R {
			if r.RPr != nil && r.RPr.RFonts != nil {
				font := r.RPr.RFonts.Ascii
				if font == nil {
					font = r.RPr.RFonts.HAnsi
				}
				if font != nil {
					lower := strings.ToLower(*font)
					if lower == "consolas" || lower == "courier" || lower == "courier new" {
						isCodeBlock = true
						break
					}
				}
			}
		}
	}

	if isCodeBlock {
		var code strings.Builder
		code.WriteString("```")
		code.WriteString("\n")
		for _, r := range ct_p.R {
			if r.T != nil {
				code.WriteString(r.T.Value)
			}
		}
		for _, hl := range ct_p.Hyperlink {
			for _, r := range hl.R {
				if r.T != nil {
					code.WriteString(r.T.Value)
				}
			}
		}
		code.WriteString("\n```")
		return code.String(), nil
	}

	var listMarker string
	if ct_p.PPr != nil && ct_p.PPr.NumPr != nil {
		marker, err := resolveListMarker(ct_p.PPr.NumPr, d)
		if err != nil {
			return "", err
		}
		listMarker = marker + " "
	}

	var body strings.Builder
	body.WriteString(prefix)
	body.WriteString(listMarker)

	body.WriteString(renderRuns(ct_p.R, ct_p.Hyperlink, d))

	result := body.String()
	if listMarker != "" {
		result = strings.TrimSpace(result)
	}
	return result, nil
}

func renderRuns(runs []*wml.CT_R, hyperlinks []*wml.CT_Hyperlink, d *Document) string {
	type formattedSegment struct {
		text   string
		format runFormatKey
	}

	var segments []formattedSegment

	for _, r := range runs {
		if r.T != nil {
			fk := runFormatKey{}
			if r.RPr != nil {
				if r.RPr.B != nil && (r.RPr.B.Val == nil || *r.RPr.B.Val) {
					fk.bold = true
				}
				if r.RPr.I != nil && (r.RPr.I.Val == nil || *r.RPr.I.Val) {
					fk.italic = true
				}
				if r.RPr.Strike != nil && (r.RPr.Strike.Val == nil || *r.RPr.Strike.Val) {
					fk.strike = true
				}
			}
			segments = append(segments, formattedSegment{text: r.T.Value, format: fk})
		}
		if r.Drawing != nil && r.Drawing.Inline != nil {
			md := resolveImage(r.Drawing.Inline, d)
			if md != "" {
				segments = append(segments, formattedSegment{text: md})
			}
		}
	}

	for _, hl := range hyperlinks {
		linkText := ""
		for _, r := range hl.R {
			if r.T != nil {
				linkText += r.T.Value
			}
		}
		url := resolveHyperlinkURL(hl.ID, d)
		if url != "" {
			md := fmt.Sprintf("[%s](%s)", linkText, url)
			segments = append(segments, formattedSegment{text: md})
		} else if linkText != "" {
			segments = append(segments, formattedSegment{text: linkText})
		}
	}

	var merged []formattedSegment
	for _, s := range segments {
		if len(merged) > 0 && merged[len(merged)-1].format == s.format {
			merged[len(merged)-1].text += s.text
		} else {
			merged = append(merged, s)
		}
	}

	var b strings.Builder
	for _, s := range merged {
		text := escapeMarkdown(s.text)
		if s.format.bold {
			text = "**" + text + "**"
		}
		if s.format.italic {
			text = "*" + text + "*"
		}
		if s.format.strike {
			text = "~~" + text + "~~"
		}
		b.WriteString(text)
	}

	return b.String()
}

func resolveListMarker(numPr *wml.CT_NumPr, d *Document) (string, error) {
	if numPr == nil || numPr.NumId == nil || numPr.NumId.Val == nil {
		return "", nil
	}

	nb := d.readOrCreateNumbering()
	numID := *numPr.NumId.Val

	var absNumID int64
	var found bool
	for _, n := range nb.Num {
		if n.NumID != nil && *n.NumID == numID {
			if n.AbstractNumID != nil && n.AbstractNumID.Val != nil {
				absNumID = *n.AbstractNumID.Val
				found = true
			}
			break
		}
	}
	if !found {
		return "", nil
	}

	var ilvl int64
	if numPr.ILvl != nil && numPr.ILvl.Val != nil {
		ilvl = *numPr.ILvl.Val
	}

	for _, an := range nb.AbstractNum {
		if an.AbstractNumID != nil && *an.AbstractNumID == absNumID {
			for _, lvl := range an.Lvl {
				if lvl.ILvl != nil && *lvl.ILvl == ilvl && lvl.NumFmt != nil && lvl.NumFmt.Val != nil {
					switch *lvl.NumFmt.Val {
					case "bullet":
						return "-", nil
					default:
						return "1", nil
					}
				}
			}
			return "1", nil
		}
	}

	return "", nil
}

func resolveImage(inline *wml.CT_Inline, d *Document) string {
	if inline.Graphic == nil || inline.Graphic.GraphicData == nil ||
		inline.Graphic.GraphicData.Pic == nil ||
		inline.Graphic.GraphicData.Pic.BlipFill == nil ||
		inline.Graphic.GraphicData.Pic.BlipFill.Blip == nil {
		return ""
	}

	rID := inline.Graphic.GraphicData.Pic.BlipFill.Blip.Embed
	if rID == "" {
		return ""
	}

	rels := d.pkg.Rels["word/document.xml"]
	if rels == nil {
		return ""
	}

	rel := findRelByID(rels, rID)
	if rel == nil {
		return ""
	}

	target := path.Join("word", rel.Target)
	part, ok := d.pkg.Parts[target]
	if !ok {
		return ""
	}

	rc, err := part.Open()
	if err != nil {
		return ""
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil || len(data) == 0 {
		return ""
	}

	mime := imageMIMEFromHeader(data)
	if mime == "" {
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	alt := ""
	if inline.DocPr != nil {
		alt = inline.DocPr.Name
	}

	return fmt.Sprintf("![%s](data:%s;base64,%s)", alt, mime, encoded)
}

func resolveHyperlinkURL(rID string, d *Document) string {
	if rID == "" {
		return ""
	}
	rels := d.pkg.Rels["word/document.xml"]
	if rels == nil {
		return ""
	}
	rel := findRelByID(rels, rID)
	if rel == nil {
		return ""
	}
	if rel.TargetMode == "External" {
		return rel.Target
	}
	return rel.Target
}

func tableToMarkdown(tbl *wml.CT_Tbl, d *Document) string {
	if tbl == nil || len(tbl.Tr) == 0 {
		return ""
	}

	var b strings.Builder

	for ri, tr := range tbl.Tr {
		var cells []string
		for _, tc := range tr.Tc {
			var cellText strings.Builder
			for _, p := range tc.P {
				text := renderRuns(p.R, p.Hyperlink, d)
				if text != "" {
					if cellText.Len() > 0 {
						cellText.WriteString(" ")
					}
					cellText.WriteString(text)
				}
			}
			cells = append(cells, escapePipeCell(cellText.String()))
		}

		b.WriteString("| ")
		b.WriteString(strings.Join(cells, " | "))
		b.WriteString(" |")
		b.WriteString("\n")

		if ri == 0 {
			b.WriteString("|")
			for range cells {
				b.WriteString(" --- |")
			}
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `*`, `\*`)
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, `[`, `\[`)
	return s
}

func escapePipeCell(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `|`, `\|`)
	return s
}

func imageMIMEFromHeader(data []byte) string {
	if len(data) < 8 {
		return ""
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0xFF && data[1] == 0xD8 {
		return "image/jpeg"
	}
	return ""
}
