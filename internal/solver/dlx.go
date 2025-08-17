package solver

import (
	"context"
	"errors"
	"time"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/ports"
)

// DLXSolver implements Algorithm X / Dancing Links for Sudoku.
// Exact-cover mapping: 324 columns (constraints), 729 rows (r,c,v candidates).
// Columns: 0..80   -> cell (r,c)
//          81..161 -> row r has number v
//          162..242-> col c has number v
//          243..323-> box b has number v, b = (r/3)*3 + (c/3)
type DLXSolver struct{}

func NewDLXSolver() *DLXSolver { return &DLXSolver{} }

const (
	nSize     = 9
	nCells    = nSize * nSize        // 81
	nCols     = 4 * nCells           // 324
	nRows     = nCells * nSize       // 729 (r,c,v)
	colCell   = 0
	colRowNum = 81
	colColNum = 162
	colBoxNum = 243
)

// node/column structures (classic dancing links)
type node struct {
	left, right, up, down *node
	col                   *column
	rowIdx                int // 0..728 identifies (r,c,v) row
}
type column struct {
	node
	size   int
	name   int
	active bool // whether this constraint column is currently uncovered
}

type dlx struct {
	cols      [nCols]*column
	rowHead   [nRows]*node
	sol       [nRows]*node
	solLen    int
	nodes     int
	activeCnt int // number of active (uncovered) columns
}

func newDLX() *dlx {
	d := &dlx{}
	// build columns
	for i := 0; i < nCols; i++ {
		c := &column{name: i, active: true}
		c.up = &c.node
		c.down = &c.node
		d.cols[i] = c
	}
	d.activeCnt = nCols

	// build rows for all (r,c,v)
	for r := 0; r < nSize; r++ {
		for c := 0; c < nSize; c++ {
			for v := 1; v <= nSize; v++ {
				row := rowIndex(r, c, v)
				cols := rowColumns(r, c, v)
				var first *node
				var prev *node
				for _, colID := range cols {
					col := d.cols[colID]
					n := &node{col: col, rowIdx: row}
					// vertical insert (at bottom)
					n.down = &col.node
					n.up = col.node.up
					col.node.up.down = n
					col.node.up = n
					col.size++
					// horizontal ring for the 4 nodes of the row
					if first == nil {
						first = n
						n.left = n
						n.right = n
					} else {
						// hook after prev
						n.left = prev
						n.right = prev.right
						prev.right.left = n
						prev.right = n
					}
					prev = n
				}
				d.rowHead[row] = first
			}
		}
	}
	return d
}

func rowIndex(r, c, v int) int {
	return (r*nSize+c)*nSize + (v - 1) // 0..728
}
func rowColumns(r, c, v int) [4]int {
	cell := colCell + r*nSize + c
	rowN := colRowNum + r*nSize + (v - 1)
	colN := colColNum + c*nSize + (v - 1)
	box := (r/3)*3 + (c / 3)
	boxN := colBoxNum + box*nSize + (v - 1)
	return [4]int{cell, rowN, colN, boxN}
}

// core operations
func cover(col *column, d *dlx) {
	if col.active {
		col.active = false
		d.activeCnt--
	}
	for i := col.down; i != &col.node; i = i.down {
		for j := i.right; j != i; j = j.right {
			j.down.up = j.up
			j.up.down = j.down
			j.col.size--
		}
	}
}
func uncover(col *column, d *dlx) {
	for i := col.up; i != &col.node; i = i.up {
		for j := i.left; j != i; j = j.left {
			j.col.size++
			j.down.up = j
			j.up.down = j
		}
	}
	if !col.active {
		col.active = true
		d.activeCnt++
	}
}

// choose the active column with the smallest size
func chooseColumn(d *dlx) *column {
	var best *column
	for _, c := range d.cols {
		if c.active {
			if best == nil || c.size < best.size {
				best = c
				if best.size == 0 {
					break
				}
			}
		}
	}
	return best
}

func (d *dlx) search(ctx context.Context, k int, wantCount int, found *int) bool {
	// cancellation check
	select {
	case <-ctx.Done():
		return true // stop search
	default:
	}
	// all constraints covered → solution
	if d.activeCnt == 0 {
		d.solLen = k
		(*found)++
		return *found >= wantCount
	}

	c := chooseColumn(d)
	if c == nil || c.size == 0 {
		return false
	}
	cover(c, d)
	for r := c.down; r != &c.node; r = r.down {
		d.nodes++
		d.sol[k] = r
		// cover other columns for this row
		for j := r.right; j != r; j = j.right {
			if j.col.active {
				cover(j.col, d)
			}
		}
		if d.search(ctx, k+1, wantCount, found) {
			// back out coverings done for this row before exiting
			for j := r.left; j != r; j = j.left {
				uncover(j.col, d)
			}
			uncover(c, d)
			return true
		}
		// backtrack: uncover in reverse order
		for j := r.left; j != r; j = j.left {
			uncover(j.col, d)
		}
	}
	uncover(c, d)
	return false
}

// apply givens by selecting corresponding rows and covering their columns
func (d *dlx) applyGiven(r, c, v int) error {
	row := rowIndex(r, c, v)
	head := d.rowHead[row]
	if head == nil {
		return errors.New("invalid row mapping")
	}
	// simulate choosing this row at top level: cover its columns
	for j := head; ; j = j.right {
		cover(j.col, d)
		if j.right == head {
			break
		}
	}
	return nil
}

func (s *DLXSolver) Solve(ctx context.Context, b *domain.Board) (*domain.Board, ports.Stats, error) {
	start := time.Now()
	d := newDLX()
	// apply givens
	for r := 0; r < nSize; r++ {
		for c := 0; c < nSize; c++ {
			if v := int(b.Values[r][c]); v > 0 {
				if v < 1 || v > 9 {
					return nil, ports.Stats{}, errors.New("invalid given")
				}
				if err := d.applyGiven(r, c, v); err != nil {
					return nil, ports.Stats{}, err
				}
			}
		}
	}
	found := 0
	_ = d.search(ctx, 0, 1, &found)
	if found < 1 {
		return nil, ports.Stats{Nodes: d.nodes, Duration: time.Since(start)}, errors.New("no solution")
	}
	// reconstruct board from chosen rows in d.sol
	var out domain.Board
	for i := 0; i < d.solLen; i++ {
		rx := d.sol[i].rowIdx
		r, c, v := decodeRow(rx)
		out.Values[r][c] = uint8(v)
	}
	return &out, ports.Stats{Nodes: d.nodes, Duration: time.Since(start)}, nil
}

func decodeRow(row int) (r, c, v int) {
	cell := row / nSize                // 0..80
	v = (row % nSize) + 1             // 1..9
	r = cell / nSize                  // 0..8
	c = cell % nSize                  // 0..8
	return
}

func (s *DLXSolver) Unique(ctx context.Context, b *domain.Board) (bool, ports.Stats, error) {
	start := time.Now()
	d := newDLX()
	for r := 0; r < nSize; r++ {
		for c := 0; c < nSize; c++ {
			if v := int(b.Values[r][c]); v > 0 {
				if v < 1 || v > 9 {
					return false, ports.Stats{}, errors.New("invalid given")
				}
				if err := d.applyGiven(r, c, v); err != nil {
					return false, ports.Stats{}, err
				}
			}
		}
	}
	found := 0
	_ = d.search(ctx, 0, 2, &found) // stop after finding 2 solutions
	unique := found == 1
	return unique, ports.Stats{Nodes: d.nodes, Duration: time.Since(start)}, nil
}