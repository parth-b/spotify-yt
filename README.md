# Spotify to YouTube Playlist Transfer

A web application that allows users to transfer their Spotify playlists to YouTube. Built with Go and modern web technologies.

## Features

- OAuth2 authentication for both Spotify and YouTube
- List all Spotify playlists
- Transfer playlists from Spotify to YouTube
- Modern, responsive UI with Tailwind CSS
- Real-time transfer status updates

## Prerequisites

- Go 1.20 or higher
- Spotify Developer Account
- Google Cloud Project with YouTube Data API v3 enabled

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/spotify-yt.git
cd spotify-yt
```

2. Create a `config.yaml` file in the root directory:
```yaml
spotify:
  client_id: "your_spotify_client_id"
  client_secret: "your_spotify_client_secret"
  redirect_uri: "http://localhost:8000/callback"

youtube:
  client_id: "your_youtube_client_id"
  client_secret: "your_youtube_client_secret"
  redirect_uri: "http://localhost:8000/youtube/callback"
```

3. Set up Spotify:
   - Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
   - Create a new application
   - Add `http://localhost:8000/callback` to the Redirect URIs

4. Set up YouTube:
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project
   - Enable YouTube Data API v3
   - Create OAuth 2.0 credentials
   - Add `http://localhost:8000/youtube/callback` to the authorized redirect URIs

5. Install dependencies:
```bash
go mod download
```

6. Run the application:
```bash
go run main.go
```

7. Visit `http://localhost:8000` in your browser

## Project Structure

```
spotify-yt/
├── pkg/
│   ├── config/     # Configuration handling
│   ├── spotify/    # Spotify API integration
│   └── youtube/    # YouTube API integration
├── static/         # Frontend assets
├── main.go         # Main application entry
├── go.mod         # Go module file
├── go.sum         # Go module checksums
└── config.yaml    # Configuration file
```

## API Endpoints

- `GET /` - Main application UI
- `GET /login` - Spotify OAuth2 login
- `GET /callback` - Spotify OAuth2 callback
- `GET /youtube/login` - YouTube OAuth2 login
- `GET /youtube/callback` - YouTube OAuth2 callback
- `GET /playlists` - List user's Spotify playlists
- `POST /transfer` - Transfer a playlist to YouTube


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 
