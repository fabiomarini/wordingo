package xmlutil

import (
	"bytes"
	"encoding/xml"
	"io"
	"sort"
	"strings"
)

// Encoder wraps xml.Encoder to emit canonical OOXML namespace
// prefixes (w, r, a, wp, …) instead of per-element default-namespace
// declarations.
//
// Usage:
//
//	var buf bytes.Buffer
//	enc := xmlutil.NewEncoder(&buf)
//	enc.Encode(myCT_P)          // two-pass: marshal internally, replay with prefix rewrite
//	enc.Flush()
//
// For token-level encoding:
//
//	enc.EncodeToken(start)      // rewrites name.Space→prefix:local
//	enc.EncodeToken(content)    // passes through unchanged
//	enc.EncodeToken(end)
type Encoder struct {
	inner  *xml.Encoder
	target io.Writer

	rootSeen     bool
	usedPrefixes map[string]struct{}
}

// NewEncoder returns an Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		target:       w,
		usedPrefixes: make(map[string]struct{}),
	}
}

// lazyInit creates the underlying xml.Encoder on first use so the
// caller can set up the Encoder without creating output immediately.
func (enc *Encoder) lazyInit() {
	if enc.inner == nil {
		enc.inner = xml.NewEncoder(enc.target)
	}
}

// EncodeToken rewrites namespace-qualified token names (StartElement
// and EndElement) to canonical prefix form before delegating to the
// underlying xml.Encoder.  On the first StartElement it also adds
// synthesized xmlns:* and mc:Ignorable attributes.
func (enc *Encoder) EncodeToken(t xml.Token) error {
	enc.lazyInit()

	switch tok := t.(type) {
	case xml.StartElement:
		se := xml.CopyToken(tok).(xml.StartElement)

		// Rewrite element name (Space→prefix:local).
		enc.rewriteName(&se.Name)

		// Rewrite namespace-qualified attribute names.
		for i := range se.Attr {
			enc.rewriteName(&se.Attr[i].Name)
		}

		// Strip all xmlns declarations (both xmlns="" and xmlns:prefix="")
		// from every element.  The root element gets canonical
		// declarations from addNSDecls; inner elements inherit.
		se.Attr = stripAllXMLNS(se.Attr)

		if !enc.rootSeen {
			enc.rootSeen = true
			enc.addNSDecls(&se)
		}
		return enc.inner.EncodeToken(se)

	case xml.EndElement:
		ee := xml.CopyToken(tok).(xml.EndElement)
		enc.rewriteName(&ee.Name)
		return enc.inner.EncodeToken(ee)

	default:
		return enc.inner.EncodeToken(t)
	}
}

// Encode marshals v to canonical-prefix XML via a two-pass approach:
// first to an internal buffer using the standard encoder, then token
// replay through EncodeToken for prefix rewriting.
func (enc *Encoder) Encode(v any) error {
	enc.lazyInit()

	var buf bytes.Buffer
	if err := xml.NewEncoder(&buf).Encode(v); err != nil {
		return err
	}
	return enc.replayAll(buf.Bytes())
}

// EncodeElement is like Encode but writes the provided start element
// explicitly before the marshaled content.
func (enc *Encoder) EncodeElement(v any, start xml.StartElement) error {
	enc.lazyInit()

	// Write the start element.
	if err := enc.EncodeToken(start); err != nil {
		return err
	}

	// Marshal inner content to a buffer, then replay only the inner
	// tokens (skip the outer wrapper that xml.Encoder.Encode adds).
	var buf bytes.Buffer
	if err := xml.NewEncoder(&buf).Encode(v); err != nil {
		return err
	}
	if err := enc.replayInner(buf.Bytes()); err != nil {
		return err
	}

	// Close the element.
	return enc.EncodeToken(xml.EndElement{Name: start.Name})
}

// Flush finalises the underlying encoder and ensures all buffered
// output reaches the target writer.
func (enc *Encoder) Flush() error {
	if enc.inner == nil {
		return nil
	}
	return enc.inner.Flush()
}

// ---- internal helpers ----

// rewriteName replaces a namespace-qualified Name (with Space set)
// with a canonical-prefix name.  Unknown namespaces are left unchanged
// (they will be emitted with xmlns="URI" by the standard encoder,
// which is semantically correct).  Returns true if the name was
// actually rewritten.
func (enc *Encoder) rewriteName(n *xml.Name) bool {
	if n.Space == "" {
		return false
	}
	if pfx := PrefixFor(n.Space); pfx != "" {
		n.Local = pfx + ":" + n.Local
		n.Space = ""
		enc.usedPrefixes[pfx] = struct{}{}
		return true
	}
	return false
}

// collectUsedPrefixes scans XML tokens and records every namespace
// prefix used in element or attribute names.  The caller uses this to
// ensure all needed xmlns:pfx declarations are emitted on the root.
func collectUsedPrefixes(data []byte) map[string]struct{} {
	used := make(map[string]struct{})
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if pfx := PrefixFor(t.Name.Space); pfx != "" {
				used[pfx] = struct{}{}
			}
			for _, a := range t.Attr {
				if pfx := PrefixFor(a.Name.Space); pfx != "" {
					used[pfx] = struct{}{}
				}
			}
		}
	}
	return used
}

// stripAllXMLNS removes all xmlns declarations from attrs — both
// plain xmlns="…" (Local=="xmlns", Space=="") and prefixed
// xmlns:pfx="…" (Space=="xmlns", Local=="pfx").
// After namespace-to-prefix rewriting these are redundant: canonical
// prefix:local form carries the same information, and only the root
// element gets synthesized xmlns:* attributes from addNSDecls.
func stripAllXMLNS(attrs []xml.Attr) []xml.Attr {
	out := make([]xml.Attr, 0, len(attrs))
	for _, a := range attrs {
		if a.Name.Local == "xmlns" && a.Name.Space == "" {
			continue // xmlns="…"
		}
		if a.Name.Space == "xmlns" {
			continue // xmlns:pfx="…"
		}
		out = append(out, a)
	}
	return out
}

// isTransitional returns true for schemas.openxmlformats.org URIs
// (the preferred conformance class for output).
func isTransitional(uri string) bool {
	return strings.Contains(uri, "schemas.openxmlformats.org")
}

// addNSDecls appends xmlns:* attributes for every prefix that was
// used by tokens seen so far, plus mc:Ignorable.
func (enc *Encoder) addNSDecls(se *xml.StartElement) {
	// Collect the reverse mapping: prefix -> canonical URI.
	// URI duplicates (Transitional + Strict → same prefix) are
	// resolved to the Transitional URI (preferred).
	prefixURI := make(map[string]string, len(CanonicalPrefixes))
	for uri, pfx := range CanonicalPrefixes {
		_, seen := prefixURI[pfx]
		if !seen || isTransitional(uri) {
			prefixURI[pfx] = uri
		}
	}
	// Override "r" to officeDocument relationships — the package
	// relationships URI also maps to "r" but must not shadow this.
	prefixURI["r"] = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"

	// Declare xmlns:* in sorted prefix order so re-encoded parts are
	// byte-deterministic (map iteration order is randomized in Go).
	used := make([]string, 0, len(enc.usedPrefixes))
	for pfx := range enc.usedPrefixes {
		used = append(used, pfx)
	}
	sort.Strings(used)
	mcDeclared := false
	for _, pfx := range used {
		if pfx == "xml" {
			continue // xml: prefix is implicit, never declared
		}
		uri, ok := prefixURI[pfx]
		if !ok {
			continue // unknown prefix — should not happen
		}
		se.Attr = append(se.Attr, xml.Attr{
			Name:  xml.Name{Local: "xmlns:" + pfx},
			Value: uri,
		})
		if pfx == "mc" {
			mcDeclared = true
		}
	}

	// mc:Ignorable for extension prefixes that have been used.
	ignorable := make([]string, 0, len(ExtensionPrefixes))
	for _, pfx := range ExtensionPrefixes {
		if _, used := enc.usedPrefixes[pfx]; used {
			ignorable = append(ignorable, pfx)
		}
	}
	if len(ignorable) > 0 {
		// The attribute name itself uses the mc prefix, so the
		// namespace must be declared or the output is not
		// well-formed XML (unbound prefix).
		if !mcDeclared {
			se.Attr = append(se.Attr, xml.Attr{
				Name:  xml.Name{Local: "xmlns:mc"},
				Value: "http://schemas.openxmlformats.org/markup-compatibility/2006",
			})
		}
		se.Attr = append(se.Attr, xml.Attr{
			Name:  xml.Name{Local: "mc:Ignorable"},
			Value: strings.Join(ignorable, " "),
		})
	}
}

// replayAll parses XML data and replays every token through
// EncodeToken.  It does a two-pass scan to discover all namespace
// prefixes before writing the root element, ensuring all xmlns:pfx
// declarations are present.
func (enc *Encoder) replayAll(data []byte) error {
	// First pass: collect all namespace prefixes before any output.
	allPrefixes := collectUsedPrefixes(data)

	// Seed usedPrefixes so addNSDecls on the root includes them.
	for pfx := range allPrefixes {
		enc.usedPrefixes[pfx] = struct{}{}
	}

	// Second pass: replay through EncodeToken which now sees the
	// full prefix set and will declare them on the root element.
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := enc.EncodeToken(tok); err != nil {
			return err
		}
	}
	return nil
}

// replayInner is like replayAll but skips the outermost element
// wrapper (start + end), replaying only the content children.
// This is used by EncodeElement where the caller already wrote the
// outer start element.
func (enc *Encoder) replayInner(data []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(data))

	// Skip the outer StartElement.
	_, err := dec.Token()
	if err != nil {
		return err
	}

	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
		if depth > 0 {
			if err := enc.EncodeToken(tok); err != nil {
				return err
			}
		}
	}
	return nil
}
