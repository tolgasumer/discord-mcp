package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Discord DiscordConfig `yaml:"discord"`
	MCP     MCPConfig     `yaml:"mcp"`
	Server  ServerConfig  `yaml:"server"`
}

// DiscordConfig holds Discord-specific configuration
type DiscordConfig struct {
	Token              string   `yaml:"token"`
	DefaultGuildID     string   `yaml:"guild_id,omitempty"`
	AllowedGuilds      []string `yaml:"allowed_guilds,omitempty"`
	MaxMessageLength   int      `yaml:"max_message_length"`
	RateLimitPerMinute int      `yaml:"rate_limit_per_minute"`
}

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	ServerName string `yaml:"server_name"`
	Version    string `yaml:"version"`
}

// ServerConfig holds general server configuration
type ServerConfig struct {
	LogLevel string `yaml:"log_level"`
	Debug    bool   `yaml:"debug"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Discord: DiscordConfig{
			Token:              "", // Must be provided by user
			MaxMessageLength:   2000,
			RateLimitPerMinute: 30,
		},
		MCP: MCPConfig{
			ServerName: "discord-mcp",
			Version:    "1.0.0",
		},
		Server: ServerConfig{
			LogLevel: "info",
			Debug:    false,
		},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filepath string) (*Config, error) {
	config := DefaultConfig()

	// If file doesn't exist, return default config
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if config.Discord.Token == "" {
		return nil, fmt.Errorf("discord.token is required")
	}

	return config, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, filepath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadFromEnv loads configuration values from environment variables
func (c *Config) LoadFromEnv() {
	if token := os.Getenv("DISCORD_TOKEN"); token != "" {
		c.Discord.Token = token
	}
	if guildID := os.Getenv("DISCORD_GUILD_ID"); guildID != "" {
		c.Discord.DefaultGuildID = guildID
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.Server.LogLevel = logLevel
	}
}
