package hint

import (
	"context"
	"fmt"

	"svw.info/sudoku/internal/domain"
)

// Singles implements a minimal Hinter that suggests naked singles.
type Singles struct{}

func NewSingles() *Singles { return &Singles{} }

// Hint returns the first found naked single if max tier allows it.
func (h *Singles) Hint(ctx context.Context, b *domain.Board, max domain.StrategyTier) (domain.Hint, bool, error) {
	if max < domain.StrategySingles {
		return domain.Hint{}, false, nil
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if b.Values[r][c] != 0 {
				continue
			}
			v, ok := soleCandidate(b, r, c)
			if ok {
				msg := fmt.Sprintf("Single: only %d fits here", v)
				return domain.Hint{
					Message:  msg,
					Cells:    []domain.CellCoord{{Row: r, Col: c}},
					Strategy: domain.StrategySingles,
				}, true, nil
			}
		}
	}
	return domain.Hint{}, false, nil
}

func soleCandidate(b *domain.Board, r, c int) (uint8, bool) {
	var last uint8
	count := 0
	for v := uint8(1); v <= 9; v++ {
		if allowed(b, r, c, v) {
			count++
			last = v
			if count > 1 {
				return 0, false
			}
		}
	}
	return last, count == 1
}

func allowed(b *domain.Board, r, c int, v uint8) bool {
	// row & col
	for i := 0; i < 9; i++ {
		if b.Values[r][i] == v || b.Values[i][c] == v {
			return false
		}
	}
	// box
	br, bc := (r/3)*3, (c/3)*3
	for dr := 0; dr < 3; dr++ {
		for dc := 0; dc < 3; dc++ {
			if b.Values[br+dr][bc+dc] == v {
				return false
			}
		}
	}
	return true
}