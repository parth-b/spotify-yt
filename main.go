package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/parth-b/spotify-yt/pkg/config"
	"github.com/parth-b/spotify-yt/pkg/spotify"
	"github.com/parth-b/spotify-yt/pkg/youtube"
)

type Server struct {
	spotifyService *spotify.Service
	youtubeService *youtube.Service
	router         *mux.Router
}

func NewServer(cfg *config.Config) (*Server, error) {
	ctx := context.Background()
	spotifyService := spotify.NewService(cfg)
	youtubeService, err := youtube.NewService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube service: %v", err)
	}

	server := &Server{
		spotifyService: spotifyService,
		youtubeService: youtubeService,
		router:         mux.NewRouter(),
	}

	server.routes()
	return server, nil
}

func (s *Server) routes() {
	// Serve static files
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Serve index.html for root path
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// Auth endpoints
	s.router.HandleFunc("/login", s.handleSpotifyLogin)
	s.router.HandleFunc("/callback", s.handleSpotifyCallback)
	s.router.HandleFunc("/youtube/login", s.handleYouTubeLogin)
	s.router.HandleFunc("/youtube/callback", s.handleYouTubeCallback)

	// API endpoints
	s.router.HandleFunc("/auth-status", s.handleAuthStatus)
	s.router.HandleFunc("/playlists", s.handleGetPlaylists)
	s.router.HandleFunc("/transfer", s.handleTransferPlaylist).Methods("POST")
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]bool{
		"spotify": s.spotifyService.IsAuthenticated(),
		"youtube": s.youtubeService.IsAuthenticated(),
	}
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleSpotifyLogin(w http.ResponseWriter, r *http.Request) {
	authURL := s.spotifyService.GetAuthURL()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *Server) handleSpotifyCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	err := s.spotifyService.CompleteAuth(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error completing auth: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) handleYouTubeLogin(w http.ResponseWriter, r *http.Request) {
	authURL := s.youtubeService.GetAuthURL()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *Server) handleYouTubeCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	err := s.youtubeService.CompleteAuth(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error completing auth: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) handleGetPlaylists(w http.ResponseWriter, r *http.Request) {
	playlists, err := s.spotifyService.GetPlaylists(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting playlists: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	type PlaylistResponse struct {
		ID     string `json:"ID"`
		Name   string `json:"Name"`
		Tracks struct {
			Total uint `json:"Total"`
		} `json:"Tracks"`
	}

	response := make([]PlaylistResponse, 0, len(playlists))
	for _, p := range playlists {
		response = append(response, PlaylistResponse{
			ID:   string(p.ID),
			Name: p.Name,
			Tracks: struct {
				Total uint `json:"Total"`
			}{
				Total: p.Tracks.Total,
			},
		})
	}

	json.NewEncoder(w).Encode(response)
}

type TransferRequest struct {
	PlaylistID string `json:"playlist_id"`
}

func (s *Server) handleTransferPlaylist(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get tracks from Spotify playlist
	tracks, err := s.spotifyService.GetPlaylistTracks(r.Context(), req.PlaylistID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting tracks: %v", err), http.StatusInternalServerError)
		return
	}

	// Create YouTube playlist
	playlist, err := s.youtubeService.CreatePlaylist(r.Context(), "Spotify Import", "Imported from Spotify")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating playlist: %v", err), http.StatusInternalServerError)
		return
	}

	// Transfer each track
	for _, track := range tracks {
		if track.Track.Track == nil {
			continue
		}

		// Search for the track on YouTube
		query := fmt.Sprintf("%s %s", track.Track.Track.Name, track.Track.Track.Artists[0].Name)
		searchResult, err := s.youtubeService.SearchVideo(r.Context(), query)
		if err != nil {
			log.Printf("Error searching for video %s: %v", query, err)
			continue
		}

		// Add the video to the playlist
		err = s.youtubeService.AddVideoToPlaylist(r.Context(), playlist.Id, searchResult.Id.VideoId)
		if err != nil {
			log.Printf("Error adding video to playlist: %v", err)
			continue
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":             "Playlist transfer completed",
		"youtube_playlist_id": playlist.Id,
	})
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Error creating server: %v", err)
	}

	log.Printf("Server starting on :8000")
	log.Fatal(http.ListenAndServe(":8000", server.router))
}
