package wordingo

import (
	"bytes"
	"path"
	"strings"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

type ScopedParts struct {
	Body    bool
	Tables  bool
	Headers bool
	Footers bool
}

type MergeOpts struct {
	ScopedParts ScopedParts
}

type span struct {
	start, end int
}

func (d *Document) Merge(data map[string]string, opts *MergeOpts) {
	if d == nil {
		panic("wordingo: Merge called on nil Document")
	}

	if opts == nil {
		opts = &MergeOpts{ScopedParts: ScopedParts{Body: true, Tables: true, Headers: true, Footers: true}}
	}

	unused := make(map[string]bool)
	for k := range data {
		unused[k] = true
	}

	if d.doc.Body == nil {
		return
	}

	if opts.ScopedParts.Body {
		for _, p := range d.doc.Body.P {
			replaceInPara(p, data, unused)
		}
	}

	if opts.ScopedParts.Tables {
		for _, tbl := range d.doc.Body.Tbl {
			replaceInTable(tbl, data, unused)
		}
	}

	if opts.ScopedParts.Headers {
		d.walkHeaderParts(func(hdr *wml.CT_Hdr) {
			for _, p := range hdr.P {
				replaceInPara(p, data, unused)
			}
		})
	}

	if opts.ScopedParts.Footers {
		d.walkFooterParts(func(ftr *wml.CT_Ftr) {
			for _, p := range ftr.P {
				replaceInPara(p, data, unused)
			}
		})
	}

	for k := range unused {
		d.warn("wordingo: merge key %q not found in document", k)
	}

	d.dirty = true
}

func replaceInPara(p *wml.CT_P, data map[string]string, unused map[string]bool) {
	if p == nil {
		return
	}
	replacePlaceholders(p.R, data, unused)
}

func replacePlaceholders(runs []*wml.CT_R, data map[string]string, unused map[string]bool) {
	var joined strings.Builder
	runSpans := make([]span, len(runs))
	for i, r := range runs {
		if r.T != nil {
			runSpans[i] = span{start: joined.Len(), end: joined.Len() + len(r.T.Value)}
			joined.WriteString(r.T.Value)
		} else {
			runSpans[i] = span{start: -1, end: -1}
		}
	}
	text := joined.String()

	pos := 0
	for {
		start := strings.Index(text[pos:], "{{")
		if start == -1 {
			break
		}
		start += pos
		endRel := strings.Index(text[start:], "}}")
		if endRel == -1 {
			break
		}
		end := start + endRel + 2
		key := text[start+2 : end-2]
		value, ok := data[key]
		if !ok {
			pos = end
			continue
		}
		delete(unused, key)

		placeholderEnd := end
		fragmentStart, fragmentEnd := -1, -1
		for i, s := range runSpans {
			if s.start == -1 {
				continue
			}
			if s.start <= start && s.end > start && fragmentStart == -1 {
				fragmentStart = i
			}
			if s.start < placeholderEnd && s.end >= placeholderEnd {
				fragmentEnd = i
			}
		}

		if fragmentStart == -1 || fragmentEnd == -1 {
			pos = placeholderEnd
			continue
		}

		if fragmentStart == fragmentEnd {
			runs[fragmentStart].T.Value = strings.Replace(runs[fragmentStart].T.Value, "{{"+key+"}}", value, 1)
		} else {
			before := runSpans[fragmentStart].start
			afterEnd := runSpans[fragmentEnd].end
			prefix := text[before:start]
			suffix := text[placeholderEnd:afterEnd]
			runs[fragmentStart].T.Value = prefix + value + suffix
			for i := fragmentStart + 1; i <= fragmentEnd; i++ {
				if runs[i].T != nil {
					runs[i].T.Value = ""
				}
			}
		}

		pos = placeholderEnd
	}
}

func replaceInTable(tbl *wml.CT_Tbl, data map[string]string, unused map[string]bool) {
	if tbl == nil {
		return
	}
	for _, tr := range tbl.Tr {
		for _, tc := range tr.Tc {
			for _, p := range tc.P {
				replaceInPara(p, data, unused)
			}
		}
	}
}

func (d *Document) walkHeaderParts(fn func(hdr *wml.CT_Hdr)) {
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

		var buf bytes.Buffer
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
		buf.WriteByte('\n')
		enc := xmlutil.NewEncoder(&buf)
		if err := enc.Encode(&hdr); err != nil {
			d.warn("wordingo: encode header part %s: %v", target, err)
			continue
		}
		if err := enc.Flush(); err != nil {
			d.warn("wordingo: flush header part %s: %v", target, err)
			continue
		}
		d.pkg.MarkModified(target, buf.Bytes())
	}
}

func (d *Document) walkFooterParts(fn func(ftr *wml.CT_Ftr)) {
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

		var buf bytes.Buffer
		buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
		buf.WriteByte('\n')
		enc := xmlutil.NewEncoder(&buf)
		if err := enc.Encode(&ftr); err != nil {
			d.warn("wordingo: encode footer part %s: %v", target, err)
			continue
		}
		if err := enc.Flush(); err != nil {
			d.warn("wordingo: flush footer part %s: %v", target, err)
			continue
		}
		d.pkg.MarkModified(target, buf.Bytes())
	}
}
