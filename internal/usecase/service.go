package usecase

import (
	"context"
	"errors"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/ports"
)

type Service struct {
	Solver    ports.Solver
	Generator ports.Generator
	Validator ports.Validator
	Hinter    ports.Hinter
	Storage   ports.Storage
}

func NewService(s ports.Solver, g ports.Generator, v ports.Validator, h ports.Hinter, st ports.Storage) *Service {
	return &Service{Solver: s, Generator: g, Validator: v, Hinter: h, Storage: st}
}

var errNotConfigured = errors.New("usecase dependency not configured")

func (u *Service) Solve(ctx context.Context, b *domain.Board) (*domain.Board, ports.Stats, error) {
	if u.Solver == nil {
		return nil, ports.Stats{}, errNotConfigured
	}
	return u.Solver.Solve(ctx, b)
}

func (u *Service) Generate(ctx context.Context, seed int64, d domain.Difficulty) (*domain.Puzzle, ports.Stats, error) {
	if u.Generator == nil {
		return nil, ports.Stats{}, errNotConfigured
	}
	return u.Generator.Generate(ctx, seed, d)
}

func (u *Service) Validate(ctx context.Context, b *domain.Board) (bool, []domain.CellCoord, error) {
	if u.Validator == nil {
		return false, nil, errNotConfigured
	}
	return u.Validator.Validate(ctx, b)
}

func (u *Service) Hint(ctx context.Context, b *domain.Board, max domain.StrategyTier) (domain.Hint, bool, error) {
	if u.Hinter == nil {
		return domain.Hint{}, false, errNotConfigured
	}
	return u.Hinter.Hint(ctx, b, max)
}

// Persistence
func (u *Service) Save(ctx context.Context, p *domain.Puzzle) error {
	if u.Storage == nil {
		return errNotConfigured
	}
	return u.Storage.Save(ctx, p)
}
func (u *Service) Load(ctx context.Context, id string) (*domain.Puzzle, error) {
	if u.Storage == nil {
		return nil, errNotConfigured
	}
	return u.Storage.Load(ctx, id)
}
func (u *Service) List(ctx context.Context) ([]domain.PuzzleMeta, error) {
	if u.Storage == nil {
		return nil, errNotConfigured
	}
	return u.Storage.List(ctx)
}