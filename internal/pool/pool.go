// Package pool fournit le cœur concurrent d'URLWatch : un worker pool borné qui
// distribue les URLs (fan-out) et collecte les résultats (fan-in) via des
// channels, en respectant l'annulation et le timeout portés par le context.
package pool

import (
	"context"
	"sync"
	"time"

	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
)

// Run vérifie toutes les urls avec au plus opts.Concurrency appels simultanés.
// Chaque vérification reçoit un context dérivé de ctx, borné par le timeout par
// URL. Toutes les URLs reçoivent un résultat (un lot annulé renvoie des résultats
// en échec) ; aucune goroutine ne fuit et tous les channels sont refermés.
func Run(ctx context.Context, c domain.Checker, urls []string, opts domain.Options) []domain.CheckResult {
	jobs := make(chan string)
	results := make(chan domain.CheckResult)

	workers := opts.Concurrency
	if workers > len(urls) {
		workers = len(urls)
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(ctx, c, opts.PerURLTimeout, jobs, results, &wg)
	}

	go fanOut(urls, jobs)

	// Fan-in : results est fermé dès que tous les workers ont terminé.
	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]domain.CheckResult, 0, len(urls))
	for r := range results {
		out = append(out, r)
	}
	return out
}

// worker consomme les URLs de jobs et émet les résultats sur results. Les types
// de channels sont directionnels pour rendre le rôle du worker explicite et sûr.
func worker(ctx context.Context, c domain.Checker, perURL time.Duration, jobs <-chan string, results chan<- domain.CheckResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for u := range jobs {
		cctx, cancel := context.WithTimeout(ctx, perURL)
		results <- c.Check(cctx, u)
		cancel()
	}
}

// fanOut émet toutes les URLs puis ferme jobs. Si le context est annulé, les
// workers rendent rapidement un résultat en échec, ce qui draine jobs sans
// blocage : toutes les URLs reçoivent donc un résultat.
func fanOut(urls []string, jobs chan<- string) {
	defer close(jobs)
	for _, u := range urls {
		jobs <- u
	}
}
