package api

import (
	"auth/internals/signup"
	"auth/internals/validation"
	"context"
	"encoding/json"
	"log/slog"
	"middleware"
	"net/http"
	"time"
)

type Server struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) Server {
	return Server{logger}
}

func (s *Server) StartServer(ctx context.Context) error {
	mux := http.NewServeMux()

	s.registerRouts(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      middleware.Wrap(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

func (s *Server) registerRouts(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/1/auth/signup", s.handleSignup)
	mux.HandleFunc("GET /api/1/auth/validate", s.validateJwt)
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	resp, err := signup.Signup()
	if err != nil {
		s.logger.ErrorContext(r.Context(), "failed to signup", "err", err)
		w.WriteHeader(500)
		return
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "failed to marshal response json", "err", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(respJson)
}

func (s *Server) validateJwt(w http.ResponseWriter, r *http.Request) {
	userID, ok := validation.Validate(r)
	if !ok {
		w.WriteHeader(403)
		return
	}

	w.WriteHeader(200)
	w.Header().Add("X-User-Id", userID)
}
