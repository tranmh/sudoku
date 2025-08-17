package solver

import (
	"context"
	"time"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/ports"
)

// Unique counts solutions up to 2 and reports whether exactly one exists.
func (s *BacktrackingSolver) Unique(ctx context.Context, b *domain.Board) (bool, ports.Stats, error) {
	start := time.Now()
	grid := b.Values
	nodes := 0
	count := 0

	var dfs func() bool
	dfs = func() bool {
		if ctx.Err() != nil || count >= 2 {
			return true // stop early
		}
		r, c, ok := findEmpty(&grid)
		if !ok {
			count++
			return count >= 2
		}
		for v := uint8(1); v <= 9; v++ {
			nodes++
			if isValid(&grid, r, c, v) {
				grid[r][c] = v
				if dfs() {
					return true
				}
				grid[r][c] = 0
			}
		}
		return false
	}
	_ = dfs()
	return count == 1, ports.Stats{Nodes: nodes, Duration: time.Since(start)}, nil
}