package xmlutil

import (
	"bytes"
	"encoding/xml"
	"io"
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
		rewrote := enc.rewriteName(&se.Name)
		if rewrote {
			// Strip the default xmlns declaration — we're using
			// canonical prefix form now.
			se.Attr = stripDefaultXMLNS(se.Attr)
		}
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

// stripDefaultXMLNS removes the default namespace declaration (an
// attribute with Local == "xmlns") from attrs.  After namespace-to-
// prefix rewriting the default xmlns is redundant because the
// prefix:x form carries the same information.
func stripDefaultXMLNS(attrs []xml.Attr) []xml.Attr {
	out := make([]xml.Attr, 0, len(attrs))
	for _, a := range attrs {
		if a.Name.Space == "" && a.Name.Local == "xmlns" {
			continue
		}
		out = append(out, a)
	}
	return out
}

// addNSDecls appends xmlns:* attributes for every prefix that was
// used by tokens seen so far, plus mc:Ignorable.
func (enc *Encoder) addNSDecls(se *xml.StartElement) {
	// Collect the reverse mapping: prefix -> canonical URI.
	// URI duplicates (Transitional + Strict → same prefix) are
	// resolved to the first Transitional entry.
	prefixURI := make(map[string]string, len(CanonicalPrefixes))
	for uri, pfx := range CanonicalPrefixes {
		if _, exists := prefixURI[pfx]; !exists {
			prefixURI[pfx] = uri
		}
	}

	for pfx := range enc.usedPrefixes {
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
	}

	// mc:Ignorable for extension prefixes that have been used.
	ignorable := make([]string, 0, len(ExtensionPrefixes))
	for _, pfx := range ExtensionPrefixes {
		if _, used := enc.usedPrefixes[pfx]; used {
			ignorable = append(ignorable, pfx)
		}
	}
	if len(ignorable) > 0 {
		se.Attr = append(se.Attr, xml.Attr{
			Name:  xml.Name{Local: "mc:Ignorable"},
			Value: strings.Join(ignorable, " "),
		})
	}
}

// replayAll parses XML data and replays every token through
// EncodeToken.
func (enc *Encoder) replayAll(data []byte) error {
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
