package domain

import "context"

// Checker vérifie une URL unique. L'implémentation par défaut effectue un réel
// appel HTTP ; un mock déterministe est utilisé dans les tests.
type Checker interface {
	Check(ctx context.Context, url string) CheckResult
}

// Store persiste et relit les lots.
type Store interface {
	Save(ctx context.Context, b Batch) error
	Get(ctx context.Context, id string) (Batch, error)
}
