package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"middleware"
	"net/http"
	"recently_played/reader"
	"time"
)

func StartServer(ctx context.Context) error {
	mux := http.NewServeMux()

	registerRouts(mux)

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

func registerRouts(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/1/recently-played/{userId}", getRecentlyPlayed)
}

func getRecentlyPlayed(writer http.ResponseWriter, req *http.Request) {
	userId := req.PathValue("userId")
	result, err := reader.GetLastPlayedTracks(req.Context(), userId)
	if err != nil {
		slog.Error("failed to get last played tracks", "err", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(result)
}
