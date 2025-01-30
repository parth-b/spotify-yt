package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	YouTubeClientID     string
	YouTubeClientSecret string
	YouTubeRedirectURI  string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		SpotifyClientID:     viper.GetString("spotify.client_id"),
		SpotifyClientSecret: viper.GetString("spotify.client_secret"),
		SpotifyRedirectURI:  viper.GetString("spotify.redirect_uri"),
		YouTubeClientID:     viper.GetString("youtube.client_id"),
		YouTubeClientSecret: viper.GetString("youtube.client_secret"),
		YouTubeRedirectURI:  viper.GetString("youtube.redirect_uri"),
	}

	return config, nil
}
