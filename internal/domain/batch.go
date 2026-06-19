package domain

import "time"

// Summarize calcule le résumé d'un ensemble de résultats pour une durée totale.
func Summarize(results []CheckResult, total time.Duration) Summary {
	s := Summary{Total: len(results), DurationMs: total.Milliseconds()}
	for _, r := range results {
		if r.OK {
			s.Up++
		} else {
			s.Down++
		}
	}
	return s
}

// NewBatch assemble un lot à partir de ses résultats et de sa durée d'exécution.
func NewBatch(id string, createdAt time.Time, results []CheckResult, d time.Duration) Batch {
	return Batch{
		ID:        id,
		CreatedAt: createdAt,
		Summary:   Summarize(results, d),
		Results:   results,
	}
}
