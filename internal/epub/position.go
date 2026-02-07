package epub

// Position represents a zero-based position in a text document.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// ByteOffsetToPosition converts a byte offset into line/character position.
// Lines and characters are zero-based.
func ByteOffsetToPosition(content []byte, offset int) Position {
	if offset < 0 {
		return Position{}
	}
	if offset > len(content) {
		offset = len(content)
	}

	line := 0
	col := 0

	for i := range offset {
		if content[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	return Position{Line: line, Character: col}
}
