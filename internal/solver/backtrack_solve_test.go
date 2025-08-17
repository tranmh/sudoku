package solver

import (
	"context"
	"testing"
	"time"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/validator"
)

// A classic, solvable Sudoku (0 = empty).
var sample = [9][9]uint8{
	{5, 3, 0, 0, 7, 0, 0, 0, 0},
	{6, 0, 0, 1, 9, 5, 0, 0, 0},
	{0, 9, 8, 0, 0, 0, 0, 6, 0},
	{8, 0, 0, 0, 6, 0, 0, 0, 3},
	{4, 0, 0, 8, 0, 3, 0, 0, 1},
	{7, 0, 0, 0, 2, 0, 0, 0, 6},
	{0, 6, 0, 0, 0, 0, 2, 8, 0},
	{0, 0, 0, 4, 1, 9, 0, 0, 5},
	{0, 0, 0, 0, 8, 0, 0, 7, 9},
}

func TestBacktrackingSolveUnder1s(t *testing.T) {
	in := &domain.Board{Values: sample}
	s := NewBacktrackingSolver()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out, st, err := s.Solve(ctx, in)
	if err != nil {
		t.Fatalf("Solve failed: %v (nodes=%d dur=%v)", err, st.Nodes, st.Duration)
	}
	// no zeros
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if out.Values[r][c] == 0 {
				t.Fatalf("unsolved cell at r=%d c=%d", r, c)
			}
		}
	}
	// valid by fast validator
	ok, conf, err := validator.New().Validate(ctx, out)
	if err != nil || !ok {
		t.Fatalf("invalid solution: err=%v conflicts=%v", err, conf)
	}
	if st.Duration > time.Second {
		t.Fatalf("took too long: %v (>1s)", st.Duration)
	}
	t.Logf("Solved in %v, nodes=%d", st.Duration, st.Nodes)
}