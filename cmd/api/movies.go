package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SemmiDev/chimovies/internal/data"
	"github.com/SemmiDev/chimovies/internal/validator"
)

func (s *app) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	err := s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = s.models.Movies.Insert(movie)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("api/v1/movies/%d", movie.ID))

	err = s.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *app) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil || id < 1 {
		s.notFoundResponse(w, r)
		return
	}

	movie, err := s.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			s.notFoundResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *app) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil {
		s.notFoundResponse(w, r)
		return
	}

	movie, err := s.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			s.notFoundResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(movie.Version), 32) != r.Header.Get("X-Expected-Version") {
			s.editConclictResponse(w, r)
			return
		}
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	err = s.readJSON(w, r, &input)
	if err != nil {
		s.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = s.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			s.editConclictResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *app) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := s.readIDParam(r)
	if err != nil {
		s.notFoundResponse(w, r)
		return
	}

	err = s.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			s.notFoundResponse(w, r)
		default:
			s.serverErrorResponse(w, r, err)
		}
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}

func (s *app) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = s.readString(qs, "title", "")
	input.Genres = s.readCSV(qs, "genres", []string{})
	input.Filters.Page = s.readInt(qs, "page", 1, v)
	input.Filters.PageSize = s.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = s.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		s.failedValidationResponse(w, r, v.Errors)
		return
	}

	movies, metadata, err := s.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		s.serverErrorResponse(w, r, err)
		return
	}

	err = s.writeJSON(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		s.serverErrorResponse(w, r, err)
	}
}
