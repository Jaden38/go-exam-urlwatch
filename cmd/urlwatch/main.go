// Command urlwatch démarre le microservice URLWatch : une API REST qui vérifie
// des lots d'URLs en parallèle (worker pool borné, context), agrège les
// résultats et les expose. Le point d'entrée câble les dépendances, configure le
// logger slog et gère l'arrêt gracieux.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Jaden38/go-exam-urlwatch/internal/api"
	"github.com/Jaden38/go-exam-urlwatch/internal/checker"
	"github.com/Jaden38/go-exam-urlwatch/internal/store"
)

func main() {
	logger := newLogger()
	slog.SetDefault(logger)

	srv := api.NewServer(checker.NewHTTP(), store.NewMemory(), logger)
	addr := envOr("ADDR", ":8080")
	httpServer := &http.Server{Addr: addr, Handler: srv.Routes()}

	// Annule le context à la réception de SIGINT/SIGTERM pour déclencher l'arrêt.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("démarrage du serveur", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("erreur du serveur", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("signal d'arrêt reçu, fermeture en cours")

	// Laisse les requêtes en cours se terminer avant de fermer.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("arrêt non gracieux", "error", err)
	}
	logger.Info("serveur arrêté proprement")
}

// newLogger construit un logger slog JSON dont le niveau provient de LOG_LEVEL.
func newLogger() *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(os.Getenv("LOG_LEVEL"))}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

// parseLevel convertit une valeur LOG_LEVEL en niveau slog (info par défaut).
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// envOr renvoie la variable d'environnement key ou def si elle est absente.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
