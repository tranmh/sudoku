package solver

// BacktrackingSolver is a straightforward recursive solver.
type BacktrackingSolver struct{}

func NewBacktrackingSolver() *BacktrackingSolver { return &BacktrackingSolver{} }

// --- helpers used by Solve/Unique (in other files) ---
func isValid(b *[9][9]uint8, r, c int, v uint8) bool {
	for i := 0; i < 9; i++ {
		if b[r][i] == v || b[i][c] == v {
			return false
		}
	}
	br, bc := (r/3)*3, (c/3)*3
	for dr := 0; dr < 3; dr++ {
		for dc := 0; dc < 3; dc++ {
			if b[br+dr][bc+dc] == v {
				return false
			}
		}
	}
	return true
}

func findEmpty(b *[9][9]uint8) (int, int, bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] == 0 {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}

// The implementations for Solve and Unique are in backtrack_solve.go and backtrack_unique.go,
// and use the helpers above.