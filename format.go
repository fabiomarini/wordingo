package wordingo

type Alignment int

const (
	AlignmentLeft   Alignment = iota
	AlignmentCenter
	AlignmentRight
	AlignmentBoth
)

func (a Alignment) String() string {
	switch a {
	case AlignmentLeft:
		return "left"
	case AlignmentCenter:
		return "center"
	case AlignmentRight:
		return "right"
	case AlignmentBoth:
		return "both"
	default:
		return ""
	}
}

type RunFormat struct {
	Bold      *bool
	Italic    *bool
	Underline *string
	Font      *string
	Size      *float64
	Color     *string
	Highlight *string
}

type ParFormat struct {
	Alignment *Alignment
	Spacing   *ParSpacing
	Indent    *ParIndent
}

type ParSpacing struct {
	Before   int64
	After    int64
	Line     int64
	LineRule string
}

type ParIndent struct {
	Left      int64
	Right     int64
	FirstLine int64
	Hanging   int64
}

// TableBorders holds table border definitions for use with TableBuilder.
type TableBorders struct {
	Top     *BorderDef
	Bottom  *BorderDef
	Left    *BorderDef
	Right   *BorderDef
	InsideH *BorderDef
	InsideV *BorderDef
}

// BorderDef defines a single border style for tables.
type BorderDef struct {
	Style string
	Size  int64
	Color string
}
