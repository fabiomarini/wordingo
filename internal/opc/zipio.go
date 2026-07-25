package opc

import (
	"archive/zip"
	"sort"
	"strings"
	"time"
)

// dosEpoch is the fixed timestamp used for deterministic output
// (MS-DOS epoch; the smallest time representable in a ZIP header).
var dosEpoch = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

// NewDeterministicHeader returns a Deflate header with a fixed
// timestamp so repeated saves produce byte-identical archives (OPC-02).
func NewDeterministicHeader(name string) *zip.FileHeader {
	h := &zip.FileHeader{Name: name, Method: zip.Deflate}
	h.Modified = dosEpoch
	return h
}

// CanonicalOrder returns part names in the ECMA-376 Part 2 §9.1.4.2
// streaming write order (OPC-02):
//
//  1. [Content_Types].xml
//  2. _rels/.rels
//  3. word/document.xml
//  4. word/_rels/document.xml.rels
//  5. remaining word/*.xml parts, each immediately followed by its
//     .rels part when present
//  6. everything else, sorted by name
func CanonicalOrder(names []string) []string {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	out := make([]string, 0, len(names))
	emit := func(n string) {
		if set[n] {
			out = append(out, n)
			delete(set, n)
		}
	}

	emit("[Content_Types].xml")
	emit("_rels/.rels")
	emit("word/document.xml")
	emit("word/_rels/document.xml.rels")

	// Remaining word/*.xml, sorted, each followed by its .rels.
	var wordParts []string
	for n := range set {
		if strings.HasPrefix(n, "word/") && strings.HasSuffix(n, ".xml") &&
			!strings.Contains(n, "/_rels/") {
			wordParts = append(wordParts, n)
		}
	}
	sort.Strings(wordParts)
	for _, n := range wordParts {
		emit(n)
		emit(relsPathFor(n))
	}

	// Everything else, sorted for determinism.
	var rest []string
	for n := range set {
		rest = append(rest, n)
	}
	sort.Strings(rest)
	out = append(out, rest...)
	return out
}
