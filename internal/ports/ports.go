package ports

import (
	"context"
	"time"

	"svw.info/sudoku/internal/domain"
)

// Stats captures performance characteristics of an operation.
type Stats struct {
	Nodes    int
	Duration time.Duration
}

// Solver solves a board and can test uniqueness.
type Solver interface {
	Solve(ctx context.Context, b *domain.Board) (*domain.Board, Stats, error)
	Unique(ctx context.Context, b *domain.Board) (bool, Stats, error)
}

// Generator creates new puzzles at a target difficulty.
type Generator interface {
	Generate(ctx context.Context, seed int64, difficulty domain.Difficulty) (*domain.Puzzle, Stats, error)
}

// Validator performs fast constraint checks (row/col/box).
type Validator interface {
	Validate(ctx context.Context, b *domain.Board) (ok bool, conflicts []domain.CellCoord, err error)
}

// Hinter returns the next logical step up to a max strategy tier.
type Hinter interface {
	Hint(ctx context.Context, b *domain.Board, max domain.StrategyTier) (domain.Hint, bool, error)
}

// Storage persists and retrieves puzzles as JSON.
type Storage interface {
	Save(ctx context.Context, p *domain.Puzzle) error
	Load(ctx context.Context, id string) (*domain.Puzzle, error)
	List(ctx context.Context) ([]domain.PuzzleMeta, error)
}