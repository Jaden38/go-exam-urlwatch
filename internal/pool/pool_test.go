package pool_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Jaden38/go-exam-urlwatch/internal/checker"
	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
	"github.com/Jaden38/go-exam-urlwatch/internal/pool"
)

func makeURLs(n int) []string {
	urls := make([]string, n)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://example.com/%d", i)
	}
	return urls
}

// TestRunRespectsConcurrency vérifie que le pool ne dépasse jamais la borne de
// concurrence demandée, mesurée par le mock.
func TestRunRespectsConcurrency(t *testing.T) {
	m := &checker.Mock{Delay: 20 * time.Millisecond}
	urls := makeURLs(50)
	opts := domain.Options{Concurrency: 5, PerURLTimeout: time.Second}

	results := pool.Run(context.Background(), m, urls, opts)

	if len(results) != len(urls) {
		t.Fatalf("results = %d, want %d", len(results), len(urls))
	}
	if got := m.MaxConcurrent(); got > opts.Concurrency {
		t.Errorf("concurrence max observée = %d, borne = %d", got, opts.Concurrency)
	}
}

// TestRunAllResultsReturned vérifie que chaque URL produit un résultat.
func TestRunAllResultsReturned(t *testing.T) {
	m := &checker.Mock{Results: map[string]domain.CheckResult{
		"https://example.com/0": {StatusCode: 200, OK: true},
		"https://example.com/1": {StatusCode: 500, OK: false, Error: "boom"},
	}}
	opts := domain.Options{Concurrency: 2, PerURLTimeout: time.Second}

	results := pool.Run(context.Background(), m, makeURLs(2), opts)
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
}

// TestRunCancellation vérifie l'interruption propre sur annulation du context :
// le traitement s'arrête vite et chaque URL reçoit tout de même un résultat.
func TestRunCancellation(t *testing.T) {
	m := &checker.Mock{Delay: 500 * time.Millisecond}
	urls := makeURLs(10)
	opts := domain.Options{Concurrency: 2, PerURLTimeout: time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	results := pool.Run(ctx, m, urls, opts)
	elapsed := time.Since(start)

	if elapsed > 400*time.Millisecond {
		t.Errorf("annulation trop lente : %v", elapsed)
	}
	if len(results) != len(urls) {
		t.Errorf("results = %d, want %d (chaque URL doit recevoir un résultat)", len(results), len(urls))
	}
	for _, r := range results {
		if r.OK {
			t.Errorf("résultat OK inattendu après annulation : %+v", r)
		}
	}
}
