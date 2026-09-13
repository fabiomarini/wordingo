package markdown

import (
	"bufio"
	"regexp"
	"strings"
)

type BlockType int

const (
	BlockParagraph BlockType = iota
	BlockHeading
	BlockCodeBlock
	BlockOrderedList
	BlockBulletList
	BlockTable
)

type Block struct {
	Type      BlockType
	Level     int
	Content   string
	Lines     []string
	Cells     [][]string
	Language  string
	ListItems []Block
	Inlines   []InlineSpan
}

type InlineSpan struct {
	Text     string
	Bold     bool
	Italic   bool
	Code     bool
	LinkURL  string
	LinkText string
	ImageURL string
	ImageAlt string
}

var (
	headingRe     = regexp.MustCompile(`^(#+)\s+(.+)$`)
	fenceOpenRe   = regexp.MustCompile("^```(\\w*)$")
	fenceCloseRe  = regexp.MustCompile("^```$")
	orderedListRe = regexp.MustCompile(`^(\s*)(\d+)\.\s+(.+)$`)
	bulletListRe  = regexp.MustCompile(`^(\s*)[-*+]\s+(.+)$`)
	tableRowRe    = regexp.MustCompile(`^\|(.+)\|$`)
	tableSepRe    = regexp.MustCompile(`^\|[-:| ]+\|?$`)

	boldRe    = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRe  = regexp.MustCompile(`\*(.+?)\*`)
	codeSpanRe = regexp.MustCompile("`([^`]+)`")
	linkRe    = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	imageRe   = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
)

func Parse(input string) ([]Block, error) {
	if input == "" {
		return nil, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(input))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var blocks []Block
	inCodeBlock := false
	codeLanguage := ""
	var codeLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if inCodeBlock {
			if fenceCloseRe.MatchString(line) {
				blocks = append(blocks, Block{
					Type:     BlockCodeBlock,
					Language: codeLanguage,
					Lines:    codeLines,
				})
				inCodeBlock = false
				codeLanguage = ""
				codeLines = nil
				continue
			}
			codeLines = append(codeLines, line)
			continue
		}

		if m := fenceOpenRe.FindStringSubmatch(line); m != nil {
			inCodeBlock = true
			codeLanguage = m[1]
			codeLines = nil
			continue
		}

		if m := headingRe.FindStringSubmatch(line); m != nil {
			block := Block{
				Type:    BlockHeading,
				Level:   len(m[1]),
				Content: m[2],
			}
			block.Inlines = parseInlines(m[2])
			blocks = append(blocks, block)
			continue
		}

		if m := tableRowRe.FindStringSubmatch(line); m != nil {
			var rows []string
			rows = append(rows, m[1])
			for i+1 < len(lines) {
				next := lines[i+1]
				if tableSepRe.MatchString(next) {
					i++
					continue
				}
				if m2 := tableRowRe.FindStringSubmatch(next); m2 != nil {
					rows = append(rows, m2[1])
					i++
					continue
				}
				break
			}
			if len(rows) >= 2 {
				var cells [][]string
				for _, row := range rows {
					parts := strings.Split(row, "|")
					cells = append(cells, parts)
				}
				blocks = append(blocks, Block{
					Type:  BlockTable,
					Cells: cells,
				})
			}
			continue
		}

		if m := orderedListRe.FindStringSubmatch(line); m != nil {
			indent := len(m[1]) / 2
			var items []Block
			items = append(items, Block{
				Type:    BlockOrderedList,
				Level:   indent,
				Content: m[3],
			})
			for i+1 < len(lines) {
				next := lines[i+1]
				if next == "" {
					break
				}
				if m2 := orderedListRe.FindStringSubmatch(next); m2 != nil {
					indent2 := len(m2[1]) / 2
					items = append(items, Block{
						Type:    BlockOrderedList,
						Level:   indent2,
						Content: m2[3],
					})
					i++
					continue
				}
				if m2 := bulletListRe.FindStringSubmatch(next); m2 != nil {
					break
				}
				break
			}
			for idx := range items {
				items[idx].Inlines = parseInlines(items[idx].Content)
			}
			blocks = append(blocks, Block{
				Type:      BlockOrderedList,
				ListItems: items,
			})
			continue
		}

		if m := bulletListRe.FindStringSubmatch(line); m != nil {
			indent := len(m[1]) / 2
			var items []Block
			items = append(items, Block{
				Type:    BlockBulletList,
				Level:   indent,
				Content: m[2],
			})
			for i+1 < len(lines) {
				next := lines[i+1]
				if next == "" {
					break
				}
				if m2 := bulletListRe.FindStringSubmatch(next); m2 != nil {
					indent2 := len(m2[1]) / 2
					items = append(items, Block{
						Type:    BlockBulletList,
						Level:   indent2,
						Content: m2[2],
					})
					i++
					continue
				}
				if m2 := orderedListRe.FindStringSubmatch(next); m2 != nil {
					break
				}
				break
			}
			for idx := range items {
				items[idx].Inlines = parseInlines(items[idx].Content)
			}
			blocks = append(blocks, Block{
				Type:      BlockBulletList,
				ListItems: items,
			})
			continue
		}

		block := Block{
			Type:    BlockParagraph,
			Content: line,
		}
		block.Inlines = parseInlines(line)
		blocks = append(blocks, block)
	}

	if inCodeBlock && len(codeLines) > 0 {
		blocks = append(blocks, Block{
			Type:     BlockCodeBlock,
			Language: codeLanguage,
			Lines:    codeLines,
		})
	}

	return blocks, nil
}

func parseInlines(text string) []InlineSpan {
	if text == "" {
		return nil
	}

	var spans []InlineSpan
	pos := 0

	for pos < len(text) {
		bestMatch := -1
		bestEnd := 0
		var bestSpan InlineSpan

		if m := codeSpanRe.FindStringSubmatchIndex(text[pos:]); m != nil {
			candidate := pos + m[0]
			if bestMatch == -1 || candidate < bestMatch {
				bestMatch = candidate
				bestEnd = pos + m[1]
				bestSpan = InlineSpan{Text: text[pos+m[2] : pos+m[3]], Code: true}
			}
		}

		if m := imageRe.FindStringSubmatchIndex(text[pos:]); m != nil {
			candidate := pos + m[0]
			if bestMatch == -1 || candidate < bestMatch {
				bestMatch = candidate
				bestEnd = pos + m[1]
				bestSpan = InlineSpan{ImageAlt: text[pos+m[2] : pos+m[3]], ImageURL: text[pos+m[4] : pos+m[5]]}
			}
		}

		if m := linkRe.FindStringSubmatchIndex(text[pos:]); m != nil {
			candidate := pos + m[0]
			if bestMatch == -1 || candidate < bestMatch {
				bestMatch = candidate
				bestEnd = pos + m[1]
				bestSpan = InlineSpan{LinkText: text[pos+m[2] : pos+m[3]], LinkURL: text[pos+m[4] : pos+m[5]]}
			}
		}

		if m := boldRe.FindStringSubmatchIndex(text[pos:]); m != nil {
			candidate := pos + m[0]
			if bestMatch == -1 || candidate < bestMatch {
				bestMatch = candidate
				bestEnd = pos + m[1]
				bestSpan = InlineSpan{Text: text[pos+m[2] : pos+m[3]], Bold: true}
			}
		}

		if m := italicRe.FindStringSubmatchIndex(text[pos:]); m != nil {
			candidate := pos + m[0]
			startChar := candidate
			endChar := pos + m[1] - 1
			prevIsStar := startChar > 0 && text[startChar-1] == '*'
			nextIsStar := endChar+1 < len(text) && text[endChar+1] == '*'
			if !prevIsStar && !nextIsStar && (bestMatch == -1 || candidate < bestMatch) {
				bestMatch = candidate
				bestEnd = pos + m[1]
				bestSpan = InlineSpan{Text: text[pos+m[2] : pos+m[3]], Italic: true}
			}
		}

		if bestMatch == -1 {
			// No inline syntax remains: the rest is literal text.
			// Emitting it as one span keeps multi-byte UTF-8 intact —
			// the old per-byte path (string(text[pos])) re-encoded
			// every byte as its own rune, producing mojibake
			// ("è" -> "Ã¨") and one run per character.
			spans = append(spans, InlineSpan{Text: text[pos:]})
			break
		}

		if bestMatch > pos {
			spans = append(spans, InlineSpan{Text: text[pos:bestMatch]})
		}
		spans = append(spans, bestSpan)
		pos = bestEnd
	}

	return spans
}
