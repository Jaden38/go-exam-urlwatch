// Package domain regroupe les types métier, les interfaces (Checker, Store) et
// les règles de validation d'URLWatch. Les autres packages dépendent de domain
// (inversion de dépendance) et jamais l'inverse.
package domain

import "time"

// CheckResult est le résultat de la vérification d'une URL.
type CheckResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	OK         bool   `json:"ok"`
	LatencyMs  int64  `json:"latency_ms"`
	Error      string `json:"error,omitempty"`
}

// Summary agrège les résultats d'un lot.
type Summary struct {
	Total      int   `json:"total"`
	Up         int   `json:"up"`
	Down       int   `json:"down"`
	DurationMs int64 `json:"duration_ms"`
}

// Batch est un lot de vérifications, persisté et relisible par son identifiant.
type Batch struct {
	ID        string        `json:"batch_id"`
	CreatedAt time.Time     `json:"created_at"`
	Summary   Summary       `json:"summary"`
	Results   []CheckResult `json:"results"`
}

// Options regroupe les paramètres d'exécution validés d'un lot.
type Options struct {
	Concurrency   int
	PerURLTimeout time.Duration
}
