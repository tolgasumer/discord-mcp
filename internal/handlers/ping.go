package handlers

import (
	"fmt"
	"time"

	"discord-mcp/internal/discord"
	"discord-mcp/pkg/types"
)

// PingTool implements a basic health check tool
type PingTool struct {
	discord *discord.Client
}

// NewPingTool creates a new ping tool
func NewPingTool(discordClient *discord.Client) *PingTool {
	return &PingTool{
		discord: discordClient,
	}
}

// Execute executes the ping tool
func (p *PingTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	startTime := time.Now()
	
	// Test Discord connection
	err := p.discord.Ping()
	if err != nil {
		return types.CallToolResult{
			Content: []types.Content{{
				Type: "text",
				Text: fmt.Sprintf("Discord connection failed: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Get bot user info
	botUser, err := p.discord.GetBotUser()
	if err != nil {
		return types.CallToolResult{
			Content: []types.Content{{
				Type: "text",
				Text: fmt.Sprintf("Failed to get bot user info: %v", err),
			}},
			IsError: true,
		}, nil
	}

	duration := time.Since(startTime)
	
	response := fmt.Sprintf("✅ Discord MCP Server is healthy!\n\n"+
		"🤖 Bot: %s#%s (ID: %s)\n"+
		"📡 Connected: %t\n"+
		"⏱️ Response time: %v\n"+
		"🕒 Timestamp: %s",
		botUser.Username,
		botUser.Discriminator,
		botUser.ID,
		p.discord.IsConnected(),
		duration,
		time.Now().Format("2006-01-02 15:04:05 UTC"))

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: response,
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (p *PingTool) GetDefinition() types.Tool {
	return types.Tool{
		Name:        "ping",
		Description: "Ping the Discord connection to verify server health and bot status",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
	}
}
