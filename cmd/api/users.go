package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/SemmiDev/chimovies/internal/data"
	"github.com/SemmiDev/chimovies/internal/validator"
)

func (s *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = s.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			s.failedValidationResponse(w, r, v.Errors)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	err = s.models.Permissions.AddForUser(user.ID, permMoviesRead, permMoviesWrite)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	token, err := s.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}
	log.Println(token.Plaintext)

	// s.background(func() {
	// 	data := map[string]interface{}{
	// 		"activationToken": token.Plaintext,
	// 		"userID":          user.ID,
	// 	}

	// 	err = s.mailer.Send(user.Email, "user_welcome.tmpl", data)
	// 	if err != nil {
	// 		s.logger.PrintError(err, nil)
	// 	}
	// })

	err = s.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := s.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			s.failedValidationResponse(w, r, v.Errors)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = s.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			s.editConclictResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	err = s.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}
