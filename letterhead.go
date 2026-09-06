package wordingo

import (
	"bytes"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/fabiomarini/wordingo/internal/opc"
	"github.com/fabiomarini/wordingo/internal/wml"
	"github.com/fabiomarini/wordingo/internal/xmlutil"
)

// letterheadCounters reports the allocation watermarks left behind by
// cloneLetterhead so callers can initialize Document.nextHeaderID,
// nextFooterID and nextImageID past the cloned parts (T-05-05: later
// AddHeader/AddImage calls must never overwrite a cloned letterhead).
type letterheadCounters struct {
	nextHeader int64
	nextFooter int64
	nextImage  int64
}

// cloneLetterhead copies the header and footer parts referenced by
// srcSectPr — plus each part's OWN relationship graph (images) and the
// media parts those relationships point at — into dst, and returns a
// sectPr carrying fresh document-level references.
//
// The part payloads are copied byte-identical; because header XML
// addresses its images through rIds inside the part's own .rels graph,
// we copy that graph too and only RENAME the media targets
// (media/imageN.png) so they cannot collide with media the caller adds
// after the clone. External relationships (TargetMode=External) are
// preserved verbatim and never fetched (D-12).
//
// OpenTemplate uses this to keep the template's body AND letterhead;
// FromTemplate uses it to keep the letterhead over an empty body.
// Shared by both so header-with-logo semantics have ONE implementation
// (the previous OpenTemplate path copied header XML but dropped its
// rels/media — a header with a logo came out with a dangling image
// reference).
func cloneLetterhead(src, dst *opc.Package, srcSectPr *wml.CT_SectPr) (*wml.CT_SectPr, letterheadCounters) {
	counters := letterheadCounters{nextHeader: 1, nextFooter: 1, nextImage: 1}
	sectPr := defaultSectPr()
	if srcSectPr == nil {
		return sectPr, counters
	}
	sectPr.TitlePg = srcSectPr.TitlePg
	if srcSectPr.PgSz != nil {
		sectPr.PgSz = srcSectPr.PgSz
	}
	if srcSectPr.PgMar != nil {
		sectPr.PgMar = srcSectPr.PgMar
	}

	srcDocRels := src.Rels["word/document.xml"]
	dstDocRels := dst.Rels["word/document.xml"]
	if dstDocRels == nil {
		dstDocRels = &opc.Relationships{}
		dst.Rels["word/document.xml"] = dstDocRels
	}

	cloneRefs := func(refs []*wml.CT_HdrFtrRef, isHeader bool) {
		for _, ref := range refs {
			srcRel := findRelByID(srcDocRels, ref.ID)
			if srcRel == nil {
				continue
			}
			target := path.Join("word", srcRel.Target)
			part, ok := src.Parts[target]
			if !ok {
				continue
			}
			partBytes, err := readPartBytes(part)
			if err != nil {
				continue
			}

			// 1. Part payload + content type + watermark.
			dst.MarkModified(target, partBytes)
			relType := relHeader
			if isHeader {
				dst.ContentTypes.Overrides["/"+target] = ctHeader
				if n := int64(partSuffix(target, "header")); n >= counters.nextHeader {
					counters.nextHeader = n + 1
				}
			} else {
				dst.ContentTypes.Overrides["/"+target] = ctFooter
				relType = relFooter
				if n := int64(partSuffix(target, "footer")); n >= counters.nextFooter {
					counters.nextFooter = n + 1
				}
			}

			// 2. The part's own relationships + their internal targets.
			clonePartRels(src, dst, target, &counters)

			// 3. Fresh document-level reference (no rId collisions).
			newID := dstDocRels.NextRID()
			dstDocRels.Rels = append(dstDocRels.Rels, opc.Relationship{
				ID:     newID,
				Type:   relType,
				Target: srcRel.Target,
			})
			sectPrRef := &wml.CT_HdrFtrRef{ID: newID, Type: ref.Type}
			if isHeader {
				sectPr.HdrFtrRef = append(sectPr.HdrFtrRef, sectPrRef)
			} else {
				sectPr.FtrRef = append(sectPr.FtrRef, sectPrRef)
			}
		}
	}
	cloneRefs(srcSectPr.HdrFtrRef, true)
	cloneRefs(srcSectPr.FtrRef, false)
	return sectPr, counters
}

// clonePartRels copies the relationship set of a cloned header/footer
// part into dst, renaming every internal media target so it cannot
// collide with media the caller adds after the clone.
func clonePartRels(src, dst *opc.Package, partName string, counters *letterheadCounters) {
	srcRels := src.Rels[partName]
	if srcRels == nil || len(srcRels.Rels) == 0 {
		return
	}
	dstRels := &opc.Relationships{}
	for _, rel := range srcRels.Rels {
		out := rel
		if out.TargetMode != "External" {
			srcTarget := path.Join("word", out.Target)
			if media, ok := src.Parts[srcTarget]; ok {
				ext := strings.ToLower(strings.TrimPrefix(path.Ext(srcTarget), "."))
				if ct := src.ContentTypes.TypeFor(srcTarget); ct != "" {
					dst.ContentTypes.Defaults[ext] = ct
				}
				if strings.HasPrefix(out.Target, "media/") {
					if data, err := readPartBytes(media); err == nil {
						newTarget := "media/image" + strconv.FormatInt(counters.nextImage, 10) + "." + ext
						counters.nextImage++
						dst.MarkModified(path.Join("word", newTarget), data)
						out.Target = newTarget
					}
				}
			}
		}
		dstRels.Rels = append(dstRels.Rels, out)
	}
	dst.Rels[partName] = dstRels
}

// partSuffix extracts the numeric suffix of a helper part name
// ("word/header2.xml", "header") -> 2; -1 when absent.
func partSuffix(partName, prefix string) int {
	base := path.Base(partName)
	if !strings.HasPrefix(base, prefix) {
		return -1
	}
	digits := strings.TrimSuffix(strings.TrimPrefix(base, prefix), ".xml")
	n, err := strconv.Atoi(digits)
	if err != nil {
		return -1
	}
	return n
}

// emptyBodyXMLWithSectPr serializes a CT_Document with no paragraphs or
// tables but a caller-supplied sectPr (the cloned letterhead references
// + page geometry). FromTemplate's empty body must still carry
// header/footer references — buildEmptyBodyXML's hardcoded defaults
// would drop them.
func emptyBodyXMLWithSectPr(sectPr *wml.CT_SectPr) []byte {
	doc := &wml.CT_Document{
		Body: &wml.CT_Body{SectPr: sectPr},
	}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteByte('\n')
	enc := xmlutil.NewEncoder(&buf)
	if err := enc.Encode(doc); err != nil {
		panic(fmt.Sprintf("wordingo: encode empty body: %v", err))
	}
	if err := enc.Flush(); err != nil {
		panic(fmt.Sprintf("wordingo: flush empty body: %v", err))
	}
	return buf.Bytes()
}
