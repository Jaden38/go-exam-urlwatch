// Package checker fournit les implémentations de domain.Checker : un
// vérificateur HTTP réel et un mock déterministe pour les tests.
package checker

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
)

// HTTP vérifie une URL par un appel GET réel. L'annulation et le timeout sont
// portés par le context passé à Check (volontairement, pas de Timeout sur le
// client) afin que le pool maîtrise le délai par URL.
type HTTP struct {
	client *http.Client
}

// NewHTTP construit un vérificateur HTTP prêt à l'emploi.
func NewHTTP() *HTTP {
	return &HTTP{client: &http.Client{}}
}

// Check effectue la requête et renvoie le résultat, en mesurant la latence même
// en cas d'échec. OK vaut vrai pour un code 2xx ou 3xx.
func (h *HTTP) Check(ctx context.Context, rawURL string) domain.CheckResult {
	start := time.Now()
	res := domain.CheckResult{URL: rawURL}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		res.Error = err.Error()
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	resp, err := h.client.Do(req)
	res.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		res.Error = classify(err)
		return res
	}
	defer resp.Body.Close()

	res.StatusCode = resp.StatusCode
	res.OK = resp.StatusCode >= 200 && resp.StatusCode < 400
	return res
}

// classify traduit les erreurs de contexte en messages courts et lisibles.
func classify(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "annulé"
	default:
		return err.Error()
	}
}
