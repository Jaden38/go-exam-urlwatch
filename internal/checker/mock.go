package checker

import (
	"context"
	"sync"
	"time"

	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
)

// Mock est un Checker déterministe pour les tests : aucun accès réseau. Il peut
// simuler une latence (annulable via le context) et mesure la concurrence
// maximale réellement atteinte, ce qui permet de vérifier la borne du pool.
type Mock struct {
	// Results associe une URL à un résultat préconfiguré. Le champ URL est
	// réécrit par Check ; les URLs absentes renvoient un succès 200 par défaut.
	Results map[string]domain.CheckResult
	// Delay simule un travail ; il est interrompu si le context est annulé.
	Delay time.Duration

	mu       sync.Mutex
	inFlight int
	maxSeen  int
}

// Check renvoie le résultat configuré pour url, en respectant l'annulation.
func (m *Mock) Check(ctx context.Context, url string) domain.CheckResult {
	m.enter()
	defer m.leave()

	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return domain.CheckResult{URL: url, Error: "annulé"}
		}
	}

	if r, ok := m.Results[url]; ok {
		r.URL = url
		return r
	}
	return domain.CheckResult{URL: url, StatusCode: 200, OK: true}
}

// MaxConcurrent renvoie le nombre maximal d'appels Check simultanés observés.
func (m *Mock) MaxConcurrent() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.maxSeen
}

func (m *Mock) enter() {
	m.mu.Lock()
	m.inFlight++
	if m.inFlight > m.maxSeen {
		m.maxSeen = m.inFlight
	}
	m.mu.Unlock()
}

func (m *Mock) leave() {
	m.mu.Lock()
	m.inFlight--
	m.mu.Unlock()
}
