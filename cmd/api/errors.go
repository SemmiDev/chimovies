package main

import (
	"fmt"
	"net/http"
)

func (s *app) logError(r *http.Request, err error) {
	s.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

func (s *app) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}
	err := s.writeJSON(w, status, env, nil)
	if err != nil {
		s.logError(r, err)
		w.WriteHeader(500)
	}
}

func (s *app) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	s.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	s.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (s *app) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	s.errorResponse(w, r, http.StatusNotFound, message)
}

func (s *app) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	s.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (s *app) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	s.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (s *app) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	s.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (s *app) editConclictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	s.errorResponse(w, r, http.StatusConflict, message)
}

func (s *app) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	s.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (s *app) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication creadentials"
	s.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (s *app) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	s.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (s *app) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resouce"
	s.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (s *app) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	s.errorResponse(w, r, http.StatusForbidden, message)
}

func (s *app) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	s.errorResponse(w, r, http.StatusForbidden, message)
}
