package epub

import "testing"

func TestPositionToByteOffset(t *testing.T) {
	content := []byte("line0\nline1\nline2")

	tests := []struct {
		name string
		pos  Position
		want int
	}{
		{"start of file", Position{Line: 0, Character: 0}, 0},
		{"middle of first line", Position{Line: 0, Character: 3}, 3},
		{"start of second line", Position{Line: 1, Character: 0}, 6},
		{"middle of second line", Position{Line: 1, Character: 2}, 8},
		{"start of third line", Position{Line: 2, Character: 0}, 12},
		{"end of file", Position{Line: 2, Character: 5}, 17},
		{"out of range line", Position{Line: 5, Character: 0}, -1},
		{"out of range character", Position{Line: 0, Character: 50}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PositionToByteOffset(content, tt.pos)
			if got != tt.want {
				t.Errorf("PositionToByteOffset(%v) = %d, want %d", tt.pos, got, tt.want)
			}
		})
	}
}

func TestPositionToByteOffset_RoundTrip(t *testing.T) {
	content := []byte("hello\nworld\nfoo")

	// For every byte offset, converting to position and back should return the same offset
	for i := range content {
		pos := ByteOffsetToPosition(content, i)
		got := PositionToByteOffset(content, pos)
		if got != i {
			t.Errorf("round-trip failed for offset %d: got %d via pos %v", i, got, pos)
		}
	}
}
