package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
	"github.com/Jaden38/go-exam-urlwatch/internal/pool"
)

// Server câble le Checker, le Store et le logger derrière les handlers HTTP.
type Server struct {
	checker domain.Checker
	store   domain.Store
	logger  *slog.Logger
}

// NewServer construit un Server.
func NewServer(c domain.Checker, s domain.Store, logger *slog.Logger) *Server {
	return &Server{checker: c, store: s, logger: logger}
}

// Routes construit le routeur et l'enveloppe des middlewares (log puis recovery).
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/checks", s.handleCreateBatch)
	mux.HandleFunc("GET /v1/checks/{id}", s.handleGetBatch)
	mux.HandleFunc("GET /healthz", s.handleHealth)

	var h http.Handler = mux
	h = JSONErrors(h)
	h = Recover(s.logger)(h)
	h = RequestLogger(s.logger)(h)
	return h
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCreateBatch(w http.ResponseWriter, r *http.Request) {
	var req checkRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "corps JSON invalide : "+err.Error())
		return
	}

	if err := domain.ValidateURLs(req.URLs); err != nil {
		s.writeDomainError(w, err)
		return
	}

	var cPtr, tPtr *int
	if req.Options != nil {
		cPtr, tPtr = req.Options.Concurrency, req.Options.TimeoutMs
	}
	opts, err := domain.NormalizeOptions(cPtr, tPtr)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), batchTimeout(opts, len(req.URLs)))
	defer cancel()

	start := time.Now()
	results := pool.Run(ctx, s.checker, req.URLs, opts)
	elapsed := time.Since(start)

	batch := domain.NewBatch(domain.NewID(), start.UTC().Truncate(time.Second), results, elapsed)
	if err := s.store.Save(r.Context(), batch); err != nil {
		s.logger.Error("persistance du lot", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "échec de persistance du lot")
		return
	}

	setBatchID(r, batch.ID)
	writeJSON(w, http.StatusCreated, batch)
}

func (s *Server) handleGetBatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batch, err := s.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrBatchNotFound) {
			writeError(w, http.StatusNotFound, "batch_not_found", fmt.Sprintf("aucun lot avec l'id %s", id))
			return
		}
		s.logger.Error("lecture du lot", "error", err)
		writeError(w, http.StatusInternalServerError, "internal", "erreur interne")
		return
	}

	setBatchID(r, id)
	writeJSON(w, http.StatusOK, batch)
}

// writeDomainError traduit une erreur métier en réponse HTTP : une
// ValidationError devient 400 invalid_request, les autres erreurs 500.
func (s *Server) writeDomainError(w http.ResponseWriter, err error) {
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		writeError(w, http.StatusBadRequest, "invalid_request", ve.Error())
		return
	}
	s.logger.Error("erreur métier inattendue", "error", err)
	writeError(w, http.StatusInternalServerError, "internal", "erreur interne")
}

// batchTimeout borne la durée totale d'un lot d'après le nombre de vagues
// d'exécution (URLs / concurrence) et le timeout par URL, plus une marge.
func batchTimeout(opts domain.Options, n int) time.Duration {
	waves := (n + opts.Concurrency - 1) / opts.Concurrency
	if waves < 1 {
		waves = 1
	}
	return time.Duration(waves)*opts.PerURLTimeout + time.Second
}
