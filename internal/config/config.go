package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	DiscordKey string
	DBPath     string
}

// LogSafe returns a map of config fields safe for logging (secrets masked).
func (c *Config) LogSafe() map[string]interface{} {
	mask := func(s string) string {
		if s == "" {
			return ""
		}
		return "***"
	}
	return map[string]interface{}{
		"db_path":     c.DBPath,
		"discord_key": mask(c.DiscordKey),
	}
}

// Load loads environment variables and returns a Config.
// If path is non-empty, it loads from that file and returns an error if the file cannot be loaded.
// If path is empty, it optionally loads .env from the current working directory; if no .env file
// exists, it does not error—DISCORD_KEY may be set in the process environment instead.
// DISCORD_KEY must be set (either from a .env file or from the environment).
func Load(path string) (*Config, error) {
	if path != "" {
		if err := godotenv.Load(path); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	} else {
		_ = godotenv.Load() // optional: ignore error if .env not present
	}

	discordKey := os.Getenv("DISCORD_KEY")
	if discordKey == "" {
		return nil, fmt.Errorf("DISCORD_KEY is not set (set it in your environment or use -env path to a .env file)")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "database.db" // Default value
	}

	return &Config{
		DiscordKey: discordKey,
		DBPath:     dbPath,
	}, nil
}
