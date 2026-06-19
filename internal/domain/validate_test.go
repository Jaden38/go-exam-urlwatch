package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Jaden38/go-exam-urlwatch/internal/domain"
)

func ptr(i int) *int { return &i }

func TestValidateURLs(t *testing.T) {
	long := make([]string, domain.MaxURLs+1)
	for i := range long {
		long[i] = "https://ok.example"
	}
	tests := []struct {
		name    string
		urls    []string
		wantErr bool
	}{
		{"vide", nil, true},
		{"trop long", long, true},
		{"http valide", []string{"http://example.com"}, false},
		{"https valide", []string{"https://example.com/x?y=1"}, false},
		{"schéma manquant", []string{"example.com"}, true},
		{"schéma interdit", []string{"ftp://example.com"}, true},
		{"hôte manquant", []string{"https://"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateURLs(tt.urls)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateURLs() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeOptions(t *testing.T) {
	tests := []struct {
		name        string
		concurrency *int
		timeoutMs   *int
		wantErr     bool
		wantConc    int
		wantTimeout time.Duration
	}{
		{"defaults", nil, nil, false, domain.DefaultConcurrency, domain.DefaultTimeoutMs * time.Millisecond},
		{"valides", ptr(4), ptr(2000), false, 4, 2000 * time.Millisecond},
		{"concurrence trop basse", ptr(0), nil, true, 0, 0},
		{"concurrence trop haute", ptr(51), nil, true, 0, 0},
		{"timeout trop bas", nil, ptr(50), true, 0, 0},
		{"timeout trop haut", nil, ptr(30001), true, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := domain.NormalizeOptions(tt.concurrency, tt.timeoutMs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if opts.Concurrency != tt.wantConc || opts.PerURLTimeout != tt.wantTimeout {
				t.Errorf("opts = %+v, want conc=%d timeout=%v", opts, tt.wantConc, tt.wantTimeout)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	results := []domain.CheckResult{{OK: true}, {OK: false}, {OK: true}}
	s := domain.Summarize(results, 1500*time.Millisecond)
	if s.Total != 3 || s.Up != 2 || s.Down != 1 || s.DurationMs != 1500 {
		t.Errorf("résumé inattendu : %+v", s)
	}
}

// TestValidationErrorAs vérifie que l'erreur de validation est détectable via
// errors.As, ce dont dépend la traduction en réponse 400 côté API.
func TestValidationErrorAs(t *testing.T) {
	err := domain.ValidateURLs(nil)
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Field != "urls" {
		t.Fatalf("attendu *ValidationError sur le champ urls, obtenu %v", err)
	}
}
