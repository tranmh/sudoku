package generator

import (
	"context"
	"math/rand"
	"time"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/ports"
)

func targetGivens(d domain.Difficulty) int {
	switch d {
	case domain.Easy:
		return 40
	case domain.Medium:
		return 34
	case domain.Hard:
		return 28
	default:
		return 24 // Expert
	}
}

// Generate creates a puzzle with a unique solution using seed and target difficulty.
func (g *UniqueGenerator) Generate(ctx context.Context, seed int64, diff domain.Difficulty) (*domain.Puzzle, ports.Stats, error) {
	start := time.Now()
	rng := rand.New(rand.NewSource(seed))
	// 1) full random solution
	var full [9][9]uint8
	if !fillRandom(ctx, rng, &full) {
		return nil, ports.Stats{}, context.Canceled
	}
	// 2) carve out clues while preserving uniqueness
	puz := full // working puzzle grid
	fixed := [9][9]bool{}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ { fixed[r][c] = true }
	}
	positions := make([]int, 81)
	for i := 0; i < 81; i++ { positions[i] = i }
	rng.Shuffle(len(positions), func(i, j int) { positions[i], positions[j] = positions[j], positions[i] })

	target := targetGivens(diff)
	deadline := start.Add(900 * time.Millisecond)
	nodes := 0

	for _, pos := range positions {
		if time.Now().After(deadline) { break }
		// stop if target reached
		if countGivens(&puz) <= target {
			break
		}
		r, c := pos/9, pos%9
		if puz[r][c] == 0 { continue }
		old := puz[r][c]
		puz[r][c] = 0
		fixed[r][c] = false
		unique, st, _ := g.Solver.Unique(ctx, &domain.Board{Values: puz})
		nodes += st.Nodes
		if !unique {
			// revert
			puz[r][c] = old
			fixed[r][c] = true
		}
	}

	p := &domain.Puzzle{
		ID:         "",
		Seed:       seed,
		Difficulty: diff,
		Board:      domain.Board{Values: puz, Fixed: fixed},
		CreatedAt:  time.Now().UnixNano(),
	}
	return p, ports.Stats{Nodes: nodes, Duration: time.Since(start)}, nil
}

func countGivens(b *[9][9]uint8) int {
	n := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b[r][c] != 0 { n++ }
		}
	}
	return n
}

// fillRandom solves an empty grid into a full valid solution by random ordering.
func fillRandom(ctx context.Context, rng *rand.Rand, grid *[9][9]uint8) bool {
	var nums [9]uint8
	for i := 0; i < 9; i++ { nums[i] = uint8(i + 1) }
	var dfs func(int, int) bool
	dfs = func(r, c int) bool {
		if ctx.Err() != nil { return false }
		if r == 9 { return true }
		nr, nc := r, c+1
		if nc == 9 { nr, nc = r+1, 0 }
		// random order
		rng.Shuffle(9, func(i, j int){ nums[i], nums[j] = nums[j], nums[i] })
		for _, v := range nums {
			if allowed(grid, r, c, v) {
				grid[r][c] = v
				if dfs(nr, nc) { return true }
				grid[r][c] = 0
			}
		}
		return false
	}
	return dfs(0,0)
}

// allowed mirrors row/col/box checks locally for the generator.
func allowed(b *[9][9]uint8, r, c int, v uint8) bool {
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