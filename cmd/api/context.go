package main

import (
	"context"
	"net/http"

	"github.com/SemmiDev/chimovies/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (s *app) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (s *app) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
