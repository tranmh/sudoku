package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"svw.info/sudoku/internal/domain"
)

type FS struct{ dir string }

func NewFS(dir string) *FS { return &FS{dir: dir} }

func diffDir(d domain.Difficulty) string {
	switch d {
	case domain.Easy:
		return "easy"
	case domain.Hard:
		return "hard"
	case domain.Expert:
		return "expert"
	default:
		return "medium"
	}
}

func (s *FS) pathFor(id string, d domain.Difficulty) string {
	sub := diffDir(d)
	return filepath.Join(s.dir, sub, strings.TrimSpace(id)+".json")
}

func (s *FS) Save(ctx context.Context, p *domain.Puzzle) error {
	if p == nil || p.ID == "" {
		return errors.New("invalid puzzle: missing ID")
	}
	// Ensure directory ./data/{difficulty} exists
	target := s.pathFor(p.ID, p.Difficulty)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

func (s *FS) Load(ctx context.Context, id string) (*domain.Puzzle, error) {
	type cand struct {
		path   string
		diff   domain.Difficulty
		legacy bool
	}
	candidates := []cand{
		{filepath.Join(s.dir, "easy", id+".json"), domain.Easy, false},
		{filepath.Join(s.dir, "medium", id+".json"), domain.Medium, false},
		{filepath.Join(s.dir, "hard", id+".json"), domain.Hard, false},
		{filepath.Join(s.dir, "expert", id+".json"), domain.Expert, false},
		{filepath.Join(s.dir, id+".json"), 0, true}, // legacy flat layout
	}
	var chosen *cand
	var data []byte
	for i := range candidates {
		c := candidates[i]
		if _, statErr := os.Stat(c.path); statErr == nil {
			b, err := os.ReadFile(c.path)
			if err != nil {
				return nil, err
			}
			data = b
			chosen = &c
			break
		}
	}
	if data == nil {
		return nil, os.ErrNotExist
	}
	var out domain.Puzzle
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	// If difficulty missing, infer from the folder we loaded from (legacy defaults to Medium)
	if out.Difficulty == 0 {
		if chosen != nil && !chosen.legacy {
			out.Difficulty = chosen.diff
		} else {
			out.Difficulty = domain.Medium
		}
	}
	return &out, nil
}

func (s *FS) List(ctx context.Context) ([]domain.PuzzleMeta, error) {
	type m struct {
		ID         string            `json:"id"`
		Name       string            `json:"name,omitempty"`
		Difficulty domain.Difficulty `json:"difficulty"`
		CreatedAt  int64             `json:"createdAt"`
	}

	var out []domain.PuzzleMeta
	// scan subfolders by difficulty
	type bucket struct {
		path string
		diff domain.Difficulty
	}
	buckets := []bucket{
		{filepath.Join(s.dir, "easy"), domain.Easy},
		{filepath.Join(s.dir, "medium"), domain.Medium},
		{filepath.Join(s.dir, "hard"), domain.Hard},
		{filepath.Join(s.dir, "expert"), domain.Expert},
	}

	for _, b := range buckets {
		ents, err := os.ReadDir(b.path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range ents {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(name, ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(b.path, name))
			if err != nil {
				continue
			}
			var mm m
			if err := json.Unmarshal(data, &mm); err != nil || mm.ID == "" {
				continue
			}
			dd := mm.Difficulty
			if dd == 0 {
				dd = b.diff // infer from folder if absent
			}
			out = append(out, domain.PuzzleMeta{
				ID:         mm.ID,
				Name:       mm.Name,
				Difficulty: dd,
				CreatedAt:  mm.CreatedAt,
			})
		}
	}

	// Also include legacy flat files in s.dir
	if ents, err := os.ReadDir(s.dir); err == nil {
		for _, e := range ents {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(name, ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(s.dir, name))
			if err != nil {
				continue
			}
			var mm m
			if err := json.Unmarshal(data, &mm); err != nil || mm.ID == "" {
				continue
			}
			dd := mm.Difficulty
			if dd == 0 {
				dd = domain.Medium
			}
			out = append(out, domain.PuzzleMeta{
				ID:         mm.ID,
				Name:       mm.Name,
				Difficulty: dd,
				CreatedAt:  mm.CreatedAt,
			})
		}
	}
	return out, nil
}