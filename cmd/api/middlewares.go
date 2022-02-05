package main

import (
	"errors"
	"expvar"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"

	"github.com/SemmiDev/chimovies/internal/data"
	"github.com/SemmiDev/chimovies/internal/validator"
)

func (s *app) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.config.LimitedEnable {
			ip := realip.FromRequest(r)
			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(s.config.LimitedRPS), s.config.LimitedBurst)}
			}
			clients[ip].lastSeen = time.Now()
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				s.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (s *app) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = s.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			s.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			s.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := s.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				s.invalidAuthenticationTokenResponse(w, r)
			default:
				s.serverErrorResponse(w, r, err)
			}
			return
		}

		r = s.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (s *app) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.contextGetUser(r)
		if user.IsAnonymous() {
			s.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *app) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.contextGetUser(r)
		if !user.Activated {
			s.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return s.requireAuthenticatedUser(fn)
}

const (
	permMoviesRead  = "movies:read"
	permMoviesWrite = "movies:write"
)

func (s *app) readPermMovies(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := s.contextGetUser(r)
		permissions, err := s.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			s.serverErrorResponse(w, r, err)
			return
		}
		if !permissions.Include(permMoviesRead) {
			s.notPermittedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
	return s.requireActivatedUser(fn)
}

func (s *app) writePermMovies(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := s.contextGetUser(r)
		permissions, err := s.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			s.serverErrorResponse(w, r, err)
			return
		}
		if !permissions.Include(permMoviesWrite) {
			s.notPermittedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
	return s.requireActivatedUser(fn)
}

func (s *app) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range s.config.TrustedOrigins {
				if origin == s.config.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *app) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_response_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalRequestsReceived.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		totalResponsesSent.Add(1)
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
