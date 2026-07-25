package opc

import (
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

// relationshipsNS is the fixed package-level relationships namespace
// (identical for Transitional and Strict packages).
const relationshipsNS = "http://schemas.openxmlformats.org/package/2006/relationships"

// Relationship is one entry of a .rels part.
type Relationship struct {
	ID         string
	Type       string
	Target     string
	TargetMode string // "" (internal) or "External"
}

// Relationships is the relationship set owned by one source part
// (or by the package root when the source part is ""). next is a
// high-water mark so deleted ids are never reused (OPC-06).
type Relationships struct {
	Rels []Relationship
	next int
}

type relsXML struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/package/2006/relationships Relationships"`
	Rels    []struct {
		ID         string `xml:"Id,attr"`
		Type       string `xml:"Type,attr"`
		Target     string `xml:"Target,attr"`
		TargetMode string `xml:"TargetMode,attr"`
	} `xml:"http://schemas.openxmlformats.org/package/2006/relationships Relationship"`
}

// parseRelationships decodes a .rels part from r.
func parseRelationships(r io.Reader) (*Relationships, error) {
	var x relsXML
	if err := xml.NewDecoder(r).Decode(&x); err != nil {
		return nil, fmt.Errorf("opc: parse relationships: %w", err)
	}
	rs := &Relationships{Rels: make([]Relationship, 0, len(x.Rels))}
	for _, rel := range x.Rels {
		rs.Rels = append(rs.Rels, Relationship{
			ID:         rel.ID,
			Type:       rel.Type,
			Target:     rel.Target,
			TargetMode: rel.TargetMode,
		})
	}
	rs.rescan()
	return rs, nil
}

// rescan raises the high-water mark above every existing rId<N>.
func (rs *Relationships) rescan() {
	for _, r := range rs.Rels {
		if n, ok := strings.CutPrefix(r.ID, "rId"); ok {
			if v, err := strconv.Atoi(n); err == nil && v >= rs.next {
				rs.next = v + 1
			}
		}
	}
}

// relsPathFor returns the .rels part name that would hold the
// relationships of the given part ("" → root "_rels/.rels").
func relsPathFor(partName string) string {
	if partName == "" {
		return "_rels/.rels"
	}
	dir, base := path.Split(partName)
	return dir + "_rels/" + base + ".rels"
}

// isRelsPath reports whether name is a .rels part.
func isRelsPath(name string) bool {
	return name == "_rels/.rels" ||
		(strings.Contains(name, "/_rels/") && strings.HasSuffix(name, ".rels"))
}

// sourcePartFor maps a .rels part name back to its source part
// ("" for the root).
func sourcePartFor(relsName string) string {
	if relsName == "_rels/.rels" {
		return ""
	}
	// "<dir>/_rels/<base>.rels" -> "<dir>/<base>"
	i := strings.LastIndex(relsName, "/_rels/")
	if i < 0 {
		return ""
	}
	dir := relsName[:i]
	base := strings.TrimSuffix(relsName[i+len("/_rels/"):], ".rels")
	if dir == "" {
		return base
	}
	return dir + "/" + base
}

// serialize renders the relationship set as a canonical .rels part.
func (rs *Relationships) serialize() ([]byte, error) {
	var x relsXML
	for _, r := range rs.Rels {
		x.Rels = append(x.Rels, struct {
			ID         string `xml:"Id,attr"`
			Type       string `xml:"Type,attr"`
			Target     string `xml:"Target,attr"`
			TargetMode string `xml:"TargetMode,attr"`
		}{r.ID, r.Type, r.Target, r.TargetMode})
	}
	out, err := xml.Marshal(x)
	if err != nil {
		return nil, fmt.Errorf("opc: serialize relationships: %w", err)
	}
	return append([]byte(xml.Header), out...), nil
}

// NextRID allocates a fresh relationship id. Ids are issued
// monotonically as rId<n> from a high-water mark; deleted ids are
// never reused (OPC-06).
func (rs *Relationships) NextRID() string {
	if rs.next == 0 {
		rs.rescan()
	}
	id := "rId" + strconv.Itoa(rs.next)
	rs.next++
	return id
}

// has reports whether id is already in use.
func (rs *Relationships) has(id string) bool {
	for _, r := range rs.Rels {
		if r.ID == id {
			return true
		}
	}
	return false
}

// validate checks the relationship graph for the source part:
// internal targets must resolve within the package, ids unique,
// targets non-empty. External relationships are preserved untouched.
// sourcePart is "" for the root set. parts holds all package part names.
func (rs *Relationships) validate(sourcePart string, parts map[string]bool) error {
	seen := make(map[string]bool, len(rs.Rels))
	for _, r := range rs.Rels {
		if r.ID == "" {
			return fmt.Errorf("opc: relationship in %s: empty Id: %w",
				relsPathFor(sourcePart), ErrInvalidPackage)
		}
		if seen[r.ID] {
			return fmt.Errorf("opc: relationship %s: duplicate Id %q: %w",
				relsPathFor(sourcePart), r.ID, ErrInvalidPackage)
		}
		seen[r.ID] = true
		if r.Target == "" {
			return fmt.Errorf("opc: relationship %s: empty Target: %w",
				r.ID, ErrInvalidPackage)
		}
		if r.TargetMode == "External" {
			continue // preserved, never fetched (T-01-04)
		}
		resolved := resolveTarget(sourcePart, r.Target)
		if resolved == "" {
			return fmt.Errorf("opc: relationship %s: target %q escapes package: %w",
				r.ID, r.Target, ErrUnsafePath)
		}
		if !parts[resolved] {
			return fmt.Errorf("opc: relationship %s: dangling target %q: %w",
				r.ID, r.Target, ErrInvalidPackage)
		}
	}
	return nil
}

// resolveTarget resolves a relative relationship target against its
// source part. Returns "" if the result escapes the package root.
func resolveTarget(sourcePart, target string) string {
	if strings.HasPrefix(target, "/") {
		// Package-absolute target.
		clean := path.Clean(target)
		if clean == "/" || strings.HasPrefix(clean, "/../") || clean == "/.." {
			return ""
		}
		return strings.TrimPrefix(clean, "/")
	}
	base := path.Dir(sourcePart)
	if sourcePart == "" {
		base = "."
	}
	joined := path.Join(base, target)
	if joined == ".." || strings.HasPrefix(joined, "../") {
		return ""
	}
	return joined
}

// sortedIDs returns relationship ids in sorted order (test aid).
func (rs *Relationships) sortedIDs() []string {
	ids := make([]string, len(rs.Rels))
	for i, r := range rs.Rels {
		ids[i] = r.ID
	}
	sort.Strings(ids)
	return ids
}
