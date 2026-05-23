package api

import (
	"catalog/internal/albums"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Response DTOs

type AlbumResponse struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	ReleaseDate *time.Time      `json:"release_date,omitempty"`
	CoverURL    *string         `json:"cover_url,omitempty"`
	Label       *string         `json:"label,omitempty"`
	TotalTracks int             `json:"total_tracks"`
	Tracks      []TrackResponse `json:"tracks,omitempty"`
}

type TrackResponse struct {
	ID          string `json:"id"`
	AlbumID     string `json:"album_id"`
	Name        string `json:"name"`
	TrackNumber int    `json:"track_number"`
	DiscNumber  int    `json:"disc_number"`
	DurationMs  int    `json:"duration_ms"`
	Explicit    bool   `json:"explicit"`
}

type ListAlbumsResponse struct {
	Albums     []AlbumResponse `json:"albums"`
	NextCursor *string         `json:"next_cursor,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Server

type Server struct {
	logger *slog.Logger
	albums *albums.Service
}

func New(logger *slog.Logger, albumsService *albums.Service) *Server {
	return &Server{
		logger: logger,
		albums: albumsService,
	}
}

func (s *Server) StartServer(ctx context.Context) error {
	mux := http.NewServeMux()

	s.registerRoutes(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
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

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.healthCheck)

	// Albums
	mux.HandleFunc("GET /api/1/albums", s.listAlbums)
	mux.HandleFunc("GET /api/1/albums/{albumID}", s.getAlbum)

	// Tracks
	mux.HandleFunc("GET /api/1/albums/{albumID}/tracks", s.listAlbumTracks)
	mux.HandleFunc("GET /api/1/tracks/{trackID}", s.getTrack)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) listAlbums(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	albumList, err := s.albums.ListAlbums(r.Context(), nil, limit)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "failed to list albums", "err", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	resp := ListAlbumsResponse{
		Albums: make([]AlbumResponse, 0, len(albumList)),
	}
	for _, a := range albumList {
		resp.Albums = append(resp.Albums, albumToResponse(&a))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) getAlbum(w http.ResponseWriter, r *http.Request) {
	albumID := r.PathValue("albumID")
	if albumID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "album id is required"})
		return
	}

	album, err := s.albums.GetAlbum(r.Context(), albumID)
	if err != nil {
		if errors.Is(err, albums.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "album not found"})
			return
		}
		s.logger.ErrorContext(r.Context(), "failed to get album", "err", err, "album_id", albumID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, albumToResponse(album))
}

func (s *Server) listAlbumTracks(w http.ResponseWriter, r *http.Request) {
	albumID := r.PathValue("albumID")
	if albumID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "album id is required"})
		return
	}

	album, err := s.albums.GetAlbum(r.Context(), albumID)
	if err != nil {
		if errors.Is(err, albums.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "album not found"})
			return
		}
		s.logger.ErrorContext(r.Context(), "failed to get album tracks", "err", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	tracks := make([]TrackResponse, 0, len(album.Tracks))
	for _, t := range album.Tracks {
		tracks = append(tracks, trackToResponse(&t))
	}

	writeJSON(w, http.StatusOK, tracks)
}

func (s *Server) getTrack(w http.ResponseWriter, r *http.Request) {
	trackID := r.PathValue("trackID")
	if trackID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "track id is required"})
		return
	}

	track, err := s.albums.GetTrack(r.Context(), trackID)
	if err != nil {
		if errors.Is(err, albums.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "track not found"})
			return
		}
		s.logger.ErrorContext(r.Context(), "failed to get track", "err", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, trackToResponse(track))
}

// Helpers

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func albumToResponse(a *albums.Album) AlbumResponse {
	resp := AlbumResponse{
		ID:          a.ID,
		Title:       a.Title,
		ReleaseDate: a.ReleaseDate,
		CoverURL:    a.CoverURL,
		Label:       a.Label,
		TotalTracks: a.TotalTracks,
	}

	if len(a.Tracks) > 0 {
		resp.Tracks = make([]TrackResponse, 0, len(a.Tracks))
		for _, t := range a.Tracks {
			resp.Tracks = append(resp.Tracks, trackToResponse(&t))
		}
	}

	return resp
}

func trackToResponse(t *albums.Track) TrackResponse {
	return TrackResponse{
		ID:          t.ID,
		AlbumID:     t.AlbumID,
		Name:        t.Name,
		TrackNumber: t.TrackNumber,
		DiscNumber:  t.DiscNumber,
		DurationMs:  t.DurationMs,
		Explicit:    t.Explicit,
	}
}
