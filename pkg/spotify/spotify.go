package spotify

import (
	"context"
	"fmt"

	"github.com/parth-b/spotify-yt/pkg/config"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Service struct {
	client *spotify.Client
	auth   *spotifyauth.Authenticator
	config *config.Config
}

func NewService(cfg *config.Config) *Service {
	auth := spotifyauth.New(
		spotifyauth.WithClientID(cfg.SpotifyClientID),
		spotifyauth.WithClientSecret(cfg.SpotifyClientSecret),
		spotifyauth.WithRedirectURL(cfg.SpotifyRedirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistReadCollaborative,
		),
	)

	return &Service{
		auth:   auth,
		config: cfg,
	}
}

func (s *Service) GetAuthURL() string {
	return s.auth.AuthURL("state")
}

func (s *Service) CompleteAuth(ctx context.Context, code string) error {
	token, err := s.auth.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("error exchanging code for token: %v", err)
	}

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)
	s.client = client
	return nil
}

func (s *Service) GetPlaylists(ctx context.Context) ([]spotify.SimplePlaylist, error) {
	if s.client == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %v", err)
	}

	playlists, err := s.client.GetPlaylistsForUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting playlists: %v", err)
	}

	return playlists.Playlists, nil
}

func (s *Service) GetPlaylistTracks(ctx context.Context, playlistID string) ([]spotify.PlaylistItem, error) {
	if s.client == nil {
		return nil, fmt.Errorf("client not authenticated")
	}

	tracks, err := s.client.GetPlaylistItems(ctx, spotify.ID(playlistID))
	if err != nil {
		return nil, fmt.Errorf("error getting playlist tracks: %v", err)
	}

	return tracks.Items, nil
}

func (s *Service) IsAuthenticated() bool {
	return s.client != nil
}
