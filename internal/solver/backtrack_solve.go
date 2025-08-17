package solver

import (
	"context"
	"errors"
	"time"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/ports"
)

func (s *BacktrackingSolver) Solve(ctx context.Context, b *domain.Board) (*domain.Board, ports.Stats, error) {
	start := time.Now()
	grid := b.Values
	nodes := 0
	var dfs func() bool
	dfs = func() bool {
		if ctx.Err() != nil {
			return false
		}
		r, c, ok := findEmpty(&grid)
		if !ok {
			return true
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
	if !dfs() {
		return nil, ports.Stats{Nodes: nodes, Duration: time.Since(start)}, errors.New("unsolvable or canceled")
	}
	out := &domain.Board{Values: grid, Fixed: b.Fixed}
	return out, ports.Stats{Nodes: nodes, Duration: time.Since(start)}, nil
}