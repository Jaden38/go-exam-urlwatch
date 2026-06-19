// Package store fournit les implémentations de domain.Store. Memory conserve les
// lots en mémoire, protégés par un RWMutex pour un accès concurrent sûr.
package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
)

// Memory est un Store en mémoire, sûr pour un usage concurrent.
type Memory struct {
	mu      sync.RWMutex
	batches map[string]domain.Batch
}

// NewMemory construit un Store en mémoire vide.
func NewMemory() *Memory {
	return &Memory{batches: make(map[string]domain.Batch)}
}

// Save enregistre (ou remplace) un lot par son identifiant.
func (m *Memory) Save(_ context.Context, b domain.Batch) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batches[b.ID] = b
	return nil
}

// Get renvoie le lot d'identifiant id, ou ErrBatchNotFound s'il n'existe pas.
func (m *Memory) Get(_ context.Context, id string) (domain.Batch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.batches[id]
	if !ok {
		return domain.Batch{}, fmt.Errorf("get %q: %w", id, domain.ErrBatchNotFound)
	}
	return b, nil
}
