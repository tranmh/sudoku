package httpadapter

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"svw.info/sudoku/internal/domain"
	"svw.info/sudoku/internal/usecase"
)

type Handler struct {
	UC *usecase.Service
}

func New(uc *usecase.Service) *Handler { return &Handler{UC: uc} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/generate", h.handleGenerate)
	mux.HandleFunc("/api/solve", h.handleSolve)
	mux.HandleFunc("/api/validate", h.handleValidate)
	mux.HandleFunc("/api/hint", h.handleHint)
	mux.HandleFunc("/api/save", h.handleSave)
	mux.HandleFunc("/api/load", h.handleLoad)
	mux.HandleFunc("/api/list", h.handleList)
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// ---- Generate ----

type generateReq struct {
	Difficulty string `json:"difficulty,omitempty"`
	Seed       int64  `json:"seed,omitempty"`
}

type generateResp struct {
	Board      domain.Board `json:"board,omitempty"`
	Seed       int64        `json:"seed,omitempty"`
	Difficulty string       `json:"difficulty,omitempty"`
	DurationMs int64        `json:"durationMs,omitempty"`
	Nodes      int          `json:"nodes,omitempty"`
	Error      string       `json:"error,omitempty"`
}

func parseDifficulty(s string) domain.Difficulty {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "easy":
		return domain.Easy
	case "hard":
		return domain.Hard
	case "expert":
		return domain.Expert
	default:
		return domain.Medium
	}
}

func (h *Handler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req generateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(generateResp{Error: "invalid JSON: " + err.Error()})
		return
	}
	seed := req.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	diff := parseDifficulty(req.Difficulty)
	p, st, err := h.UC.Generate(r.Context(), seed, diff)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(generateResp{Error: err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(generateResp{
		Board:      p.Board,
		Seed:       seed,
		Difficulty: req.Difficulty,
		DurationMs: st.Duration.Milliseconds(),
		Nodes:      st.Nodes,
	})
}

// ---- Validate ----

type validateReq struct {
	Board [9][9]uint8 `json:"board"`
}
type validateResp struct {
	OK        bool               `json:"ok"`
	Conflicts []domain.CellCoord `json:"conflicts,omitempty"`
	Error     string             `json:"error,omitempty"`
}

func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req validateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(validateResp{Error: "invalid JSON: " + err.Error()})
		return
	}
	b := &domain.Board{Values: req.Board}
	ok, conflicts, err := h.UC.Validate(r.Context(), b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(validateResp{Error: err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(validateResp{OK: ok, Conflicts: conflicts})
}

// ---- Solve ----

type solveReq struct {
	Board [9][9]uint8 `json:"board"`
}
type solveResp struct {
	Board      [9][9]uint8 `json:"board,omitempty"`
	DurationMs int64       `json:"durationMs,omitempty"`
	Nodes      int         `json:"nodes,omitempty"`
	Error      string      `json:"error,omitempty"`
}

func (h *Handler) handleSolve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req solveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(solveResp{Error: "invalid JSON: " + err.Error()})
		return
	}
	in := &domain.Board{Values: req.Board}
	out, st, err := h.UC.Solve(r.Context(), in)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(solveResp{Error: err.Error(), DurationMs: st.Duration.Milliseconds(), Nodes: st.Nodes})
		return
	}
	_ = json.NewEncoder(w).Encode(solveResp{Board: out.Values, DurationMs: st.Duration.Milliseconds(), Nodes: st.Nodes})
}

// ---- Hint ----

type hintReq struct {
	Board   [9][9]uint8 `json:"board"`
	MaxTier string      `json:"maxTier,omitempty"`
}
type hintResp struct {
	Found bool         `json:"found"`
	Hint  domain.Hint  `json:"hint,omitempty"`
	Error string       `json:"error,omitempty"`
}

func parseTier(s string) domain.StrategyTier {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "singles":
		return domain.StrategySingles
	case "pairs":
		return domain.StrategyPairs
	case "advanced":
		return domain.StrategyAdvanced
	case "xwing":
		return domain.StrategyXWing
	default:
		return domain.StrategySingles
	}
}

func (h *Handler) handleHint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req hintReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(hintResp{Error: "invalid JSON: " + err.Error()})
		return
	}
	max := parseTier(req.MaxTier)
	b := &domain.Board{Values: req.Board}
	hh, ok, err := h.UC.Hint(r.Context(), b, max)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(hintResp{Error: err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(hintResp{Found: ok, Hint: hh})
}

// ---- Save / Load / List ----

type saveResp struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func (h *Handler) handleSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var p domain.Puzzle
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(saveResp{Error: "invalid JSON: " + err.Error()})
		return
	}
	if p.ID == "" {
		p.ID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	if p.CreatedAt == 0 {
		p.CreatedAt = time.Now().UnixNano()
	}
	if err := h.UC.Save(r.Context(), &p); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(saveResp{Error: err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(saveResp{ID: p.ID})
}

type loadReq struct {
	ID string `json:"id"`
}
type loadResp struct {
	Puzzle *domain.Puzzle `json:"puzzle,omitempty"`
	Error  string         `json:"error,omitempty"`
}

func (h *Handler) handleLoad(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req loadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(loadResp{Error: "invalid JSON or missing id"})
		return
	}
	p, err := h.UC.Load(r.Context(), req.ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(loadResp{Error: err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(loadResp{Puzzle: p})
}

type listResp struct {
	Puzzles []domain.PuzzleMeta `json:"puzzles"`
	Error   string              `json:"error,omitempty"`
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	ps, err := h.UC.List(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(listResp{Error: err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(listResp{Puzzles: ps})
}