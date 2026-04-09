package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"ai-for-oj/internal/config"
)

type ShutdownHook func(ctx context.Context) error

type App struct {
	config        config.Config
	logger        *slog.Logger
	server        *http.Server
	shutdownHooks []ShutdownHook
}

func NewApp(cfg config.Config, logger *slog.Logger, handler http.Handler, shutdownHooks ...ShutdownHook) *App {
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	return &App{
		config: cfg,
		logger: logger,
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
		shutdownHooks: shutdownHooks,
	}
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		a.logger.Info("http server starting",
			"addr", a.server.Addr,
			"env", a.config.App.Env,
			"service", a.config.App.Name,
		)

		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}

		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	for _, hook := range a.shutdownHooks {
		if err := hook(shutdownCtx); err != nil {
			a.logger.Error("shutdown hook failed", "error", err)
		}
	}

	a.logger.Info("application stopped cleanly")
	return nil
}
