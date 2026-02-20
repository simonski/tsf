package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const DefaultServerURL = "http://localhost:8080"

type Config struct {
	ServerURL string
	Username  string
	Password  string
	JSON      bool
	Verbose   bool
	ProjectID string
}

type SavedCredentials struct {
	ServerURL string `json:"server_url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type SessionToken struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
	ServerURL    string `json:"server_url"`
	Username     string `json:"username"`
}

func NewConfig(args []string) (*Config, error) {
	cfg := &Config{
		ServerURL: os.Getenv("SF_URL"),
		Username:  os.Getenv("SF_USERNAME"),
		Password:  os.Getenv("SF_PASSWORD"),
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = DefaultServerURL
	}

	// Load saved credentials if not provided via env
	if cfg.Username == "" || cfg.Password == "" {
		if creds, err := LoadCredentials(); err == nil {
			if cfg.Username == "" {
				cfg.Username = creds.Username
			}
			if cfg.Password == "" {
				cfg.Password = creds.Password
			}
			if cfg.ServerURL == DefaultServerURL && creds.ServerURL != "" {
				cfg.ServerURL = creds.ServerURL
			}
		}
	}

	// Parse global flags. -json/-v/-url are safe to parse anywhere.
	// -username/-password are parsed only from leading global section to avoid
	// colliding with subcommand flags like `sf user create -username ...`.
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-url":
			if i+1 < len(args) {
				cfg.ServerURL = args[i+1]
				i++
			}
		case "-json":
			cfg.JSON = true
		case "-v":
			cfg.Verbose = true
		}
	}
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") {
			break
		}
		switch args[i] {
		case "-username":
			if i+1 < len(args) {
				cfg.Username = args[i+1]
				i++
			}
		case "-password":
			if i+1 < len(args) {
				cfg.Password = args[i+1]
				i++
			}
		case "-url":
			if i+1 < len(args) {
				cfg.ServerURL = args[i+1]
				i++
			}
		}
	}
	if projectID, err := loadProjectContext(); err == nil {
		cfg.ProjectID = projectID
	}
	return cfg, nil
}

func loadProjectContext() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(home, ".config", "sf", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	var config struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}
	return config.ProjectID, nil
}

func SaveProjectContext(projectID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "sf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	config := struct {
		ProjectID string `json:"project_id"`
	}{ProjectID: projectID}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	return os.WriteFile(configPath, data, 0644)
}

func ClearProjectContext() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".config", "sf", "config.json")
	return os.Remove(configPath)
}

func LoadCredentials() (*SavedCredentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	credPath := filepath.Join(home, ".config", "sf", "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, err
	}
	var creds SavedCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}

func SaveCredentials(serverURL, username, password string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "sf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	creds := SavedCredentials{
		ServerURL: serverURL,
		Username:  username,
		Password:  password,
	}
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	credPath := filepath.Join(configDir, "credentials.json")
	return os.WriteFile(credPath, data, 0600) // 0600 for security - only owner can read/write
}

func LoadSessionToken() (*SessionToken, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sessionPath := filepath.Join(home, ".config", "sf", "session.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}
	var session SessionToken
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func SaveSessionToken(token, refreshToken, expiresAt, serverURL, username string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "sf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	session := SessionToken{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		ServerURL:    serverURL,
		Username:     username,
	}
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	sessionPath := filepath.Join(configDir, "session.json")
	return os.WriteFile(sessionPath, data, 0600)
}

func ClearSessionToken() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sessionPath := filepath.Join(home, ".config", "sf", "session.json")
	return os.Remove(sessionPath)
}

func ClearCredentials() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	credPath := filepath.Join(home, ".config", "sf", "credentials.json")
	return os.Remove(credPath)
}
