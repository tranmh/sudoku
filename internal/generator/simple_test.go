package generator

import (
	"context"
	"testing"
	"time"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/solver"
)

func TestGenerateAllDifficultiesUnder1s(t *testing.T) {
	s := solver.NewBacktrackingSolver()
	g := NewUniqueGenerator(s)

	cases := []struct {
		name string
		diff domain.Difficulty
	}{
		{"easy", domain.Easy},
		{"medium", domain.Medium},
		{"hard", domain.Hard},
		{"expert", domain.Expert},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			seed := int64(12345)
			p, st, err := g.Generate(ctx, seed, tc.diff)
			if err != nil {
				t.Fatalf("Generate(%s) failed: %v", tc.name, err)
			}
			if st.Duration > time.Second {
				t.Fatalf("generation too slow for %s: %v (>1s)", tc.name, st.Duration)
			}
			// basic sanity: count givens (should be at least a valid baseline)
			givens := 0
			for r := 0; r < 9; r++ {
				for c := 0; c < 9; c++ {
					if p.Board.Values[r][c] != 0 {
						givens++
					}
				}
			}
			if givens < 17 || givens > 81 {
				t.Fatalf("invalid givens count for %s: %d", tc.name, givens)
			}
			// verify uniqueness
			ok, _, _ := s.Unique(ctx, &p.Board)
			if !ok {
				t.Fatalf("puzzle for %s is not unique", tc.name)
			}
		})
	}
}