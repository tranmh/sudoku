package domain

// Board holds current values and which cells are fixed givens.
type Board struct {
	Values [9][9]uint8 `json:"board"`
	Fixed  [9][9]bool  `json:"fixed,omitempty"`
}

// CellCoord identifies a cell on the board.
type CellCoord struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// Hint describes a strategy suggestion for the UI.
type Hint struct {
	Message  string       `json:"message,omitempty"`
	Cells    []CellCoord  `json:"cells,omitempty"`
	Strategy StrategyTier `json:"strategy,omitempty"`
}

// Puzzle is a persisted Sudoku with metadata.
type Puzzle struct {
	ID         string     `json:"id,omitempty"`
	Seed       int64      `json:"seed,omitempty"`
	Difficulty Difficulty `json:"difficulty,omitempty"`
	Board      Board      `json:"board"`
	CreatedAt  int64      `json:"createdAt,omitempty"`
	// Optional user metadata
	Name  string `json:"name,omitempty"`
	Notes string `json:"notes,omitempty"`
}

// PuzzleMeta is a lightweight listing entry.
type PuzzleMeta struct {
	ID         string     `json:"id"`
	Name       string     `json:"name,omitempty"`
	Difficulty Difficulty `json:"difficulty"`
	CreatedAt  int64      `json:"createdAt"`
}