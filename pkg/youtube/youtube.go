package youtube

import (
	"context"
	"fmt"

	"github.com/parth-b/spotify-yt/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Service struct {
	service *youtube.Service
	config  *config.Config
	oauth   *oauth2.Config
}

func NewService(ctx context.Context, cfg *config.Config) (*Service, error) {
	oauth := &oauth2.Config{
		ClientID:     cfg.YouTubeClientID,
		ClientSecret: cfg.YouTubeClientSecret,
		RedirectURL:  cfg.YouTubeRedirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/youtube.force-ssl",
			"https://www.googleapis.com/auth/youtube.upload",
			"https://www.googleapis.com/auth/youtube.readonly",
		},
		Endpoint: google.Endpoint,
	}

	return &Service{
		config: cfg,
		oauth:  oauth,
	}, nil
}

func (s *Service) GetAuthURL() string {
	return s.oauth.AuthCodeURL("state")
}

func (s *Service) CompleteAuth(ctx context.Context, code string) error {
	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("error exchanging code for token: %v", err)
	}

	client := s.oauth.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("error creating YouTube service: %v", err)
	}

	s.service = service
	return nil
}

func (s *Service) SearchVideo(ctx context.Context, query string) (*youtube.SearchResult, error) {
	if s.service == nil {
		return nil, fmt.Errorf("YouTube service not authenticated")
	}

	call := s.service.Search.List([]string{"id", "snippet"}).
		Q(query).
		Type("video").
		MaxResults(1)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error searching for video: %v", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no videos found for query: %s", query)
	}

	return response.Items[0], nil
}

func (s *Service) CreatePlaylist(ctx context.Context, title string, description string) (*youtube.Playlist, error) {
	if s.service == nil {
		return nil, fmt.Errorf("YouTube service not authenticated")
	}

	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       title,
			Description: description,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "private",
		},
	}

	call := s.service.Playlists.Insert([]string{"snippet", "status"}, playlist)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error creating playlist: %v", err)
	}

	return response, nil
}

func (s *Service) AddVideoToPlaylist(ctx context.Context, playlistID string, videoID string) error {
	if s.service == nil {
		return fmt.Errorf("YouTube service not authenticated")
	}

	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}

	call := s.service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err := call.Do()
	if err != nil {
		return fmt.Errorf("error adding video to playlist: %v", err)
	}

	return nil
}

func (s *Service) IsAuthenticated() bool {
	return s.service != nil
}
