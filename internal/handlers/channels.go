package handlers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"discord-mcp/internal/discord"
	"discord-mcp/internal/permissions"
	"discord-mcp/internal/validation"
	"discord-mcp/pkg/types"
)

// ChannelHandler handles Discord channel operations
type ChannelHandler struct {
	discord     *discord.Client
	permissions *permissions.Checker
	validator   *validation.Validator
	logger      *logrus.Logger
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(discordClient *discord.Client, permChecker *permissions.Checker, validator *validation.Validator, logger *logrus.Logger) *ChannelHandler {
	return &ChannelHandler{
		discord:     discordClient,
		permissions: permChecker,
		validator:   validator,
		logger:      logger,
	}
}

// ListChannelsTool implements the list_channels MCP tool
type ListChannelsTool struct {
	handler *ChannelHandler
}

// NewListChannelsTool creates a new list channels tool
func NewListChannelsTool(handler *ChannelHandler) *ListChannelsTool {
	return &ListChannelsTool{handler: handler}
}

// Execute executes the list_channels tool
func (t *ListChannelsTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("list_channels", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)

	var filterType string
	if typeVal, ok := params.Arguments["type"]; ok {
		filterType = typeVal.(string)
	}

	var includePerms bool
	if permsVal, ok := params.Arguments["include_permissions"]; ok {
		includePerms = permsVal.(bool)
	}

	// Validate permissions
	if err := t.handler.permissions.CanViewGuild(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get channels from Discord
	channels, err := t.handler.discord.GetChannels(guildID)
	if err != nil {
		return t.formatError("Failed to list channels", err), nil
	}

	// Filter channels
	filteredChannels := t.filterChannels(channels, filterType)

	// Format channels for response
	formattedChannels := make([]map[string]interface{}, len(filteredChannels))
	for i, ch := range filteredChannels {
		formattedChannels[i] = t.formatChannel(ch, includePerms)
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Found %d channels in guild %s", len(formattedChannels), guildID),
			Data: map[string]interface{}{
				"guild_id":      guildID,
				"channel_count": len(formattedChannels),
				"channels":      formattedChannels,
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *ListChannelsTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("list_channels", "List channels in a Discord server (guild)")
}

// filterChannels filters channels by type
func (t *ListChannelsTool) filterChannels(channels []*discordgo.Channel, filterType string) []*discordgo.Channel {
	if filterType == "" {
		return channels
	}

	var filtered []*discordgo.Channel
	for _, ch := range channels {
		if strings.EqualFold(channelTypeToString(ch.Type), filterType) {
			filtered = append(filtered, ch)
		}
	}
	return filtered
}

func channelTypeToString(channelType discordgo.ChannelType) string {
	switch channelType {
	case discordgo.ChannelTypeGuildText:
		return "text"
	case discordgo.ChannelTypeGuildVoice:
		return "voice"
	case discordgo.ChannelTypeGuildCategory:
		return "category"
	case discordgo.ChannelTypeGuildNews:
		return "news"
	case discordgo.ChannelTypeGuildStore:
		return "store"
	case discordgo.ChannelTypeGuildNewsThread:
		return "news_thread"
	case discordgo.ChannelTypeGuildPublicThread:
		return "public_thread"
	case discordgo.ChannelTypeGuildPrivateThread:
		return "private_thread"
	default:
		return "unknown"
	}
}

// formatChannel formats a single channel for the response
func (t *ListChannelsTool) formatChannel(channel *discordgo.Channel, includePerms bool) map[string]interface{} {
	data := map[string]interface{}{
		"id":         channel.ID,
		"name":       channel.Name,
		"type":       channelTypeToString(channel.Type),
		"position":   channel.Position,
		"nsfw":       channel.NSFW,
		"parent_id":  channel.ParentID,
		"guild_id":   channel.GuildID,
		"topic":      channel.Topic,
		"created_at": channel.LastMessageID, // This is not correct, but there is no direct created_at field
	}

	if includePerms {
		perms, err := t.handler.permissions.GetChannelPermissions(channel.ID)
		if err != nil {
			t.handler.logger.Warnf("Could not get permissions for channel %s: %v", channel.ID, err)
			data["permissions"] = "error"
		} else {
			data["permissions"] = perms
		}
	}

	return data
}

// formatError creates a standardized error response
func (t *ListChannelsTool) formatError(message string, err error) types.CallToolResult {
	t.handler.logger.Errorf("%s: %v", message, err)
	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("❌ %s: %v", message, err),
			Data: map[string]interface{}{
				"error_type": "discord_api",
				"message":    message,
				"details":    err.Error(),
			},
		}},
		IsError: true,
	}
}

// GetChannelInfoTool implements the get_channel_info MCP tool
type GetChannelInfoTool struct {
	handler *ChannelHandler
}

// NewGetChannelInfoTool creates a new get channel info tool
func NewGetChannelInfoTool(handler *ChannelHandler) *GetChannelInfoTool {
	return &GetChannelInfoTool{handler: handler}
}

// Execute executes the get_channel_info tool
func (t *GetChannelInfoTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("get_channel_info", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	channelID := params.Arguments["channel_id"].(string)

	var includePerms bool
	if permsVal, ok := params.Arguments["include_permissions"]; ok {
		includePerms = permsVal.(bool)
	}

	// Validate permissions
	if err := t.handler.permissions.CanViewChannel(channelID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get channel info from Discord
	channel, err := t.handler.discord.Session().Channel(channelID)
	if err != nil {
		return t.formatError("Failed to get channel info", err), nil
	}

	// Format channel for response
	formattedChannel := t.formatChannel(channel, includePerms)

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Channel: %s", channel.Name),
			Data: formattedChannel,
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *GetChannelInfoTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("get_channel_info", "Get information about a specific Discord channel")
}

// formatChannel formats a single channel for the response
func (t *GetChannelInfoTool) formatChannel(channel *discordgo.Channel, includePerms bool) map[string]interface{} {
	data := map[string]interface{}{
		"id":         channel.ID,
		"name":       channel.Name,
		"type":       channelTypeToString(channel.Type),
		"position":   channel.Position,
		"nsfw":       channel.NSFW,
		"parent_id":  channel.ParentID,
		"guild_id":   channel.GuildID,
		"topic":      channel.Topic,
		"created_at": channel.LastMessageID, // This is not correct, but there is no direct created_at field
	}

	if includePerms {
		perms, err := t.handler.permissions.GetChannelPermissions(channel.ID)
		if err != nil {
			t.handler.logger.Warnf("Could not get permissions for channel %s: %v", channel.ID, err)
			data["permissions"] = "error"
		} else {
			data["permissions"] = perms
		}
	}

	return data
}

// formatError creates a standardized error response
func (t *GetChannelInfoTool) formatError(message string, err error) types.CallToolResult {
	t.handler.logger.Errorf("%s: %v", message, err)
	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("❌ %s: %v", message, err),
			Data: map[string]interface{}{
				"error_type": "discord_api",
				"message":    message,
				"details":    err.Error(),
			},
		}},
		IsError: true,
	}
}