package main

import (
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	httpadapter "svw.info/sudoku/internal/adapters/http"
	"svw.info/sudoku/internal/generator"
	"svw.info/sudoku/internal/infrastructure/storage"
	"svw.info/sudoku/internal/ports"
	"svw.info/sudoku/internal/solver"
	"svw.info/sudoku/internal/usecase"
	"svw.info/sudoku/internal/validator"
	"svw.info/sudoku/internal/hint"
	"svw.info/sudoku/web"
)

// statusWriter captures HTTP status and bytes written.
type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// requestLogger logs method, path, status, bytes, and duration in a human-readable format.
func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		dur := time.Since(start)
		logger.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"bytes", sw.bytes,
			"dur", dur.Round(time.Millisecond),
		)
	})
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	persist := flag.String("persist-path", "./data", "save directory")
	levelStr := flag.String("log-level", "info", "debug|info|warn|error")
	solverKind := flag.String("solver", "dlx", "solver to use: dlx|backtrack")
	flag.Parse()

	lvl := slog.LevelInfo
	switch strings.ToLower(*levelStr) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
	_ = os.MkdirAll(*persist, 0o755)

	// Choose solver: DLX by default, backtracking as fallback via flag.
	var s ports.Solver
	switch strings.ToLower(strings.TrimSpace(*solverKind)) {
	case "backtrack", "backtracking":
		s = solver.NewBacktrackingSolver()
	default:
		s = solver.NewDLXSolver()
	}

	// Wire providers → use cases → HTTP adapter
	g := generator.NewUniqueGenerator(s)
	v := validator.New()
	st := storage.NewFS(*persist)
	hin := hint.NewSingles()
	uc := usecase.NewService(s, g, v, hin, st)
	h := httpadapter.New(uc)

	tmpl := web.Templates()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(web.StaticFS())))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.tmpl", map[string]any{}); err != nil {
			http.Error(w, template.HTMLEscapeString(err.Error()), http.StatusInternalServerError)
		}
	})
	h.Register(mux)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           requestLogger(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("listening", "addr", *addr, "persist", *persist, "solver", *solverKind)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}