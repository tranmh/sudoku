package validator

import (
	"context"

	"svw.info/sudoku/internal/domain"
)

type FastValidator struct{}

func New() *FastValidator { return &FastValidator{} }

func (v *FastValidator) Validate(ctx context.Context, b *domain.Board) (bool, []domain.CellCoord, error) {
	conf := make([]domain.CellCoord, 0, 8)
	// rows
	for r := 0; r < 9; r++ {
		m := 0
		for c := 0; c < 9; c++ {
			val := b.Values[r][c]
			if val == 0 {
				continue
			}
			bit := 1 << val
			if m&bit != 0 {
				conf = append(conf, domain.CellCoord{Row: r, Col: c})
			}
			m |= bit
		}
	}
	// cols
	for c := 0; c < 9; c++ {
		m := 0
		for r := 0; r < 9; r++ {
			val := b.Values[r][c]
			if val == 0 {
				continue
			}
			bit := 1 << val
			if m&bit != 0 {
				conf = append(conf, domain.CellCoord{Row: r, Col: c})
			}
			m |= bit
		}
	}
	// boxes
	for br := 0; br < 3; br++ {
		for bc := 0; bc < 3; bc++ {
			m := 0
			for dr := 0; dr < 3; dr++ {
				for dc := 0; dc < 3; dc++ {
					r := br*3 + dr
					c := bc*3 + dc
					val := b.Values[r][c]
					if val == 0 {
						continue
					}
					bit := 1 << val
					if m&bit != 0 {
						conf = append(conf, domain.CellCoord{Row: r, Col: c})
					}
					m |= bit
				}
			}
		}
	}
	return len(conf) == 0, conf, nil
}