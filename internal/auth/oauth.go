package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scope        string
}

var GoogleConfig = OAuthConfig{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURI:  "http://localhost:8080/auth/google/callback",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://oauth2.googleapis.com/token",
	UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
	Scope:        "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile",
}

var GitHubConfig = OAuthConfig{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	RedirectURI:  "http://localhost:8080/auth/github/callback",
	AuthURL:      "https://github.com/login/oauth/authorize",
	TokenURL:     "https://github.com/login/oauth/access_token",
	UserInfoURL:  "https://api.github.com/user",
	Scope:        "user:email",
}

func (c *OAuthConfig) GetAuthURL(state string) string {
	return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		c.AuthURL, c.ClientID, c.RedirectURI, c.Scope, state)
}

func (c *OAuthConfig) Exchange(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", c.RedirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", c.TokenURL, nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = data.Encode()
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("oauth token exchange failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no access token found in response")
	}

	return token, nil
}

type UserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`    // Google
	AvatarURL string `json:"avatar_url"` // GitHub
	Login     string `json:"login"`      // GitHub username
}

func (c *OAuthConfig) GetUserInfo(token string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", c.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info: %s", resp.Status)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// Normalize fields
	if userInfo.Name == "" && userInfo.Login != "" {
		userInfo.Name = userInfo.Login
	}
	if userInfo.Picture == "" && userInfo.AvatarURL != "" {
		userInfo.Picture = userInfo.AvatarURL
	}

	return &userInfo, nil
}
