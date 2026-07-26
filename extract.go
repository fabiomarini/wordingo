package wordingo

import (
	"bytes"
	"path"
	"strings"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

type ExtractOpts struct {
	ScopedParts ScopedParts
	Separator   string
}

func paragraphText(ct_p *wml.CT_P) string {
	if ct_p == nil {
		return ""
	}
	var b strings.Builder
	for _, r := range ct_p.R {
		if r.T != nil {
			b.WriteString(r.T.Value)
		}
	}
	for _, hl := range ct_p.Hyperlink {
		for _, r := range hl.R {
			if r.T != nil {
				b.WriteString(r.T.Value)
			}
		}
	}
	return b.String()
}

func (d *Document) readHeaderParts(fn func(hdr *wml.CT_Hdr)) {
	if d.doc.Body.SectPr == nil {
		return
	}
	rels := d.pkg.Rels["word/document.xml"]
	if rels == nil {
		return
	}
	for _, ref := range d.doc.Body.SectPr.HdrFtrRef {
		rel := findRelByID(rels, ref.ID)
		if rel == nil {
			continue
		}
		target := path.Join("word", rel.Target)
		part, ok := d.pkg.Parts[target]
		if !ok {
			continue
		}
		partBytes, err := readPartBytes(part)
		if err != nil {
			continue
		}
		var hdr wml.CT_Hdr
		dec := xmlutil.NewSafeDecoder(bytes.NewReader(partBytes), opc.MaxPartBytes)
		if err := dec.Decode(&hdr); err != nil {
			continue
		}
		fn(&hdr)
	}
}

func (d *Document) readFooterParts(fn func(ftr *wml.CT_Ftr)) {
	if d.doc.Body.SectPr == nil {
		return
	}
	rels := d.pkg.Rels["word/document.xml"]
	if rels == nil {
		return
	}
	for _, ref := range d.doc.Body.SectPr.FtrRef {
		rel := findRelByID(rels, ref.ID)
		if rel == nil {
			continue
		}
		target := path.Join("word", rel.Target)
		part, ok := d.pkg.Parts[target]
		if !ok {
			continue
		}
		partBytes, err := readPartBytes(part)
		if err != nil {
			continue
		}
		var ftr wml.CT_Ftr
		dec := xmlutil.NewSafeDecoder(bytes.NewReader(partBytes), opc.MaxPartBytes)
		if err := dec.Decode(&ftr); err != nil {
			continue
		}
		fn(&ftr)
	}
}

func (d *Document) ExtractText(opts *ExtractOpts) (string, error) {
	if d == nil {
		panic("wordingo: ExtractText called on nil Document")
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
					text := paragraphText(el.Para.X())
					if text != "" {
						b.WriteString(text)
						b.WriteString(sep)
					}
				}
			case ElementTable:
				if opts.ScopedParts.Tables {
					tbl := el.Table.X()
					for _, tr := range tbl.Tr {
						for ci, tc := range tr.Tc {
							for _, p := range tc.P {
								text := paragraphText(p)
								if text != "" {
									b.WriteString(text)
								}
							}
							if ci < len(tr.Tc)-1 {
								b.WriteString(sep)
							}
						}
						b.WriteString(sep)
					}
				}
			}
		}
	}

	if opts.ScopedParts.Headers {
		d.readHeaderParts(func(hdr *wml.CT_Hdr) {
			for _, p := range hdr.P {
				text := paragraphText(p)
				if text != "" {
					b.WriteString(text)
					b.WriteString(sep)
				}
			}
		})
	}

	if opts.ScopedParts.Footers {
		d.readFooterParts(func(ftr *wml.CT_Ftr) {
			for _, p := range ftr.P {
				text := paragraphText(p)
				if text != "" {
					b.WriteString(text)
					b.WriteString(sep)
				}
			}
		})
	}

	result := strings.TrimSuffix(b.String(), sep)
	return result, nil
}
