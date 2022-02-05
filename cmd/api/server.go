package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (s *app) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.ServerAddress),
		Handler:      s.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		q := <-quit

		s.logger.PrintInfo("shutting down server", map[string]string{
			"signal": q.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	s.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  s.config.Environment,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	s.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
