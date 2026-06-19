package domain

import (
	"fmt"
	"net/url"
	"time"
)

// Bornes et valeurs par défaut du contrat d'API (cf. README).
const (
	MaxURLs = 100

	DefaultConcurrency = 8
	MinConcurrency     = 1
	MaxConcurrency     = 50

	DefaultTimeoutMs = 5000
	MinTimeoutMs     = 100
	MaxTimeoutMs     = 30000
)

// ValidateURLs vérifie que la liste contient de 1 à MaxURLs entrées, chacune
// étant une URL http/https valide.
func ValidateURLs(urls []string) error {
	if len(urls) < 1 || len(urls) > MaxURLs {
		return &ValidationError{Field: "urls", Message: fmt.Sprintf("doit contenir entre 1 et %d entrées", MaxURLs)}
	}
	for i, raw := range urls {
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return &ValidationError{Field: fmt.Sprintf("urls[%d]", i), Message: "doit être une URL http ou https valide"}
		}
	}
	return nil
}

// NormalizeOptions applique les valeurs par défaut aux champs absents (pointeur
// nil) puis valide les bornes. Elle renvoie des Options prêtes à l'emploi.
func NormalizeOptions(concurrency, timeoutMs *int) (Options, error) {
	c := DefaultConcurrency
	if concurrency != nil {
		c = *concurrency
	}
	if c < MinConcurrency || c > MaxConcurrency {
		return Options{}, &ValidationError{Field: "options.concurrency", Message: fmt.Sprintf("doit être compris entre %d et %d", MinConcurrency, MaxConcurrency)}
	}

	t := DefaultTimeoutMs
	if timeoutMs != nil {
		t = *timeoutMs
	}
	if t < MinTimeoutMs || t > MaxTimeoutMs {
		return Options{}, &ValidationError{Field: "options.timeout_ms", Message: fmt.Sprintf("doit être compris entre %d et %d", MinTimeoutMs, MaxTimeoutMs)}
	}

	return Options{Concurrency: c, PerURLTimeout: time.Duration(t) * time.Millisecond}, nil
}
