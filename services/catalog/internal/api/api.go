package api

import (
	"context"
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
	mux.HandleFunc("GET /api/1/albums", s.getAlbums)
	mux.HandleFunc("GET /api/1/album/{albumID}", s.getAlbumInfo)
}

func (s *Server) getAlbums(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) getAlbumInfo(w http.ResponseWriter, r *http.Request) {

}

