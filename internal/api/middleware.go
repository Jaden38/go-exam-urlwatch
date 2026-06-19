package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey int

const batchIDKey ctxKey = iota

// batchIDHolder permet à un handler de remonter l'identifiant de lot au
// middleware de log, qui ne le connaît pas a priori.
type batchIDHolder struct{ id string }

// setBatchID associe un batch_id à la requête courante pour le journal d'accès.
func setBatchID(r *http.Request, id string) {
	if h, ok := r.Context().Value(batchIDKey).(*batchIDHolder); ok {
		h.id = id
	}
}

// statusRecorder capture le code de statut écrit par le handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// RequestLogger journalise chaque requête (method, path, status, duration_ms et
// batch_id si connu). La sonde /healthz est ignorée pour ne pas polluer les logs.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			holder := &batchIDHolder{}
			r = r.WithContext(context.WithValue(r.Context(), batchIDKey, holder))
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			start := time.Now()
			next.ServeHTTP(rec, r)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			}
			if holder.id != "" {
				attrs = append(attrs, "batch_id", holder.id)
			}
			logger.Info("requête traitée", attrs...)
		})
	}
}

// jsonErrorWriter convertit les erreurs de routage générées par le ServeMux
// (404 chemin inconnu, 405 méthode non autorisée) en corps d'erreur JSON. Les
// réponses applicatives, qui posent déjà Content-Type: application/json, passent
// inchangées.
type jsonErrorWriter struct {
	http.ResponseWriter
	swallow bool
}

func (w *jsonErrorWriter) WriteHeader(status int) {
	if (status == http.StatusNotFound || status == http.StatusMethodNotAllowed) &&
		w.Header().Get("Content-Type") != "application/json" {
		w.swallow = true
		code, message := routingErrorBody(status)
		writeError(w.ResponseWriter, status, code, message)
		return
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *jsonErrorWriter) Write(b []byte) (int, error) {
	if w.swallow {
		// Le corps texte par défaut du ServeMux est ignoré : le JSON est déjà écrit.
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func routingErrorBody(status int) (code, message string) {
	if status == http.StatusMethodNotAllowed {
		return "method_not_allowed", "méthode non autorisée"
	}
	return "not_found", "ressource introuvable"
}

// JSONErrors rend les erreurs de routage du ServeMux conformes au contrat
// d'erreur JSON de l'API.
func JSONErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&jsonErrorWriter{ResponseWriter: w}, r)
	})
}

// Recover transforme toute panic en réponse 500 propre et la journalise. Il est
// placé au plus près du handler pour que le statut 500 soit connu du logger.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic récupérée", "error", fmt.Sprint(rec), "path", r.URL.Path)
					writeError(w, http.StatusInternalServerError, "internal", "erreur interne")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
