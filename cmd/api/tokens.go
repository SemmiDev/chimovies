package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/SemmiDev/chimovies/internal/data"
	"github.com/SemmiDev/chimovies/internal/validator"
)

func (s *app) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := s.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			s.invalidCredentialsResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		s.invalidCredentialsResponse(w, r)
		return
	}

	token, err := s.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	err = s.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}
