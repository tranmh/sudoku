package generator

import "svw.info/sudoku/internal/ports"

// UniqueGenerator creates puzzles with a unique solution using a provided Solver.
type UniqueGenerator struct {
	Solver ports.Solver
}

// NewUniqueGenerator wires a generator that uses the given solver for uniqueness checks.
func NewUniqueGenerator(s ports.Solver) *UniqueGenerator {
	return &UniqueGenerator{Solver: s}
}

// Note: The Generate method is implemented in simple.go to avoid duplicate definitions.