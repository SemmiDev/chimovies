package main

import (
	"expvar"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *app) routes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(s.metrics)
	router.Use(s.enableCORS)
	router.Use(s.rateLimit)
	router.Use(s.authenticate)

	router.NotFound(s.notFoundResponse)
	router.MethodNotAllowed(s.methodNotAllowedResponse)

	router.Get("/api/v1/healthcheck", s.healthcheckHandler)
	router.With(s.readPermMovies).Get("/api/v1/movies", s.listMoviesHandler)
	router.With(s.writePermMovies).Post("/api/v1/movies", s.createMovieHandler)
	router.With(s.readPermMovies).Get("/api/v1/movies/{id}", s.showMovieHandler)
	router.With(s.writePermMovies).Patch("/api/v1/movies/{id}", s.updateMovieHandler)
	router.With(s.writePermMovies).Delete("/api/v1/movies/{id}", s.deleteMovieHandler)

	router.Post("/api/v1/users", s.registerUserHandler)
	router.Put("/api/v1/users/activated", s.activateUserHandler)
	router.Post("/api/v1/tokens/authentication", s.createAuthenticationTokenHandler)

	router.Get("/api/debug/vars", expvar.Handler().ServeHTTP)
	return router
}
