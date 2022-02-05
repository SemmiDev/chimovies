package main

import (
	"net/http"
)

func (s *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": s.config.Environment,
			"version":     s.config.Version,
		},
	}

	err := s.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}
