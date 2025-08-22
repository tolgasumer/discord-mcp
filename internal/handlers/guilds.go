package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"discord-mcp/internal/discord"
	"discord-mcp/internal/permissions"
	"discord-mcp/internal/validation"
	"discord-mcp/pkg/types"
)

// GuildHandler handles Discord guild operations
type GuildHandler struct {
	discord     *discord.Client
	permissions *permissions.Checker
	validator   *validation.Validator
	logger      *logrus.Logger
}

// NewGuildHandler creates a new guild handler
func NewGuildHandler(discordClient *discord.Client, permChecker *permissions.Checker, validator *validation.Validator, logger *logrus.Logger) *GuildHandler {
	return &GuildHandler{
		discord:     discordClient,
		permissions: permChecker,
		validator:   validator,
		logger:      logger,
	}
}

// GetGuildInfoTool implements the get_guild_info MCP tool
type GetGuildInfoTool struct {
	handler *GuildHandler
}

// NewGetGuildInfoTool creates a new get guild info tool
func NewGetGuildInfoTool(handler *GuildHandler) *GetGuildInfoTool {
	return &GetGuildInfoTool{handler: handler}
}

// Execute executes the get_guild_info tool
func (t *GetGuildInfoTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("get_guild_info", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanViewGuild(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get guild from Discord
	guild, err := t.handler.discord.GetGuild(guildID)
	if err != nil {
		return t.formatError("Failed to get guild info", err), nil
	}

	// Format guild for response
	formattedGuild := t.formatGuild(guild)

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Guild: %s", guild.Name),
			Data: formattedGuild,
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *GetGuildInfoTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("get_guild_info", "Get information about a specific Discord server (guild)")
}

// formatGuild formats a single guild for the response
func (t *GetGuildInfoTool) formatGuild(guild *discordgo.Guild) map[string]interface{} {
	return map[string]interface{}{
		"id":          guild.ID,
		"name":        guild.Name,
		"description": guild.Description,
		"icon":        guild.Icon,
		"splash":      guild.Splash,
		"banner":      guild.Banner,
		"owner_id":    guild.OwnerID,
		"member_count": guild.MemberCount,
	}
}

// formatError creates a standardized error response
func (t *GetGuildInfoTool) formatError(message string, err error) types.CallToolResult {
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

// ListGuildMembersTool implements the list_guild_members MCP tool
type ListGuildMembersTool struct {
	handler *GuildHandler
}

// NewListGuildMembersTool creates a new list guild members tool
func NewListGuildMembersTool(handler *GuildHandler) *ListGuildMembersTool {
	return &ListGuildMembersTool{handler: handler}
}

// Execute executes the list_guild_members tool
func (t *ListGuildMembersTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("list_guild_members", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanViewGuild(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get members from Discord
	members, err := t.handler.discord.Session().GuildMembers(guildID, "", 1000)
	if err != nil {
		return t.formatError("Failed to list guild members", err), nil
	}

	// Format members for response
	formattedMembers := make([]map[string]interface{}, len(members))
	for i, member := range members {
		formattedMembers[i] = t.formatMember(member)
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Found %d members in guild %s", len(formattedMembers), guildID),
			Data: map[string]interface{}{
				"guild_id":      guildID,
				"member_count":  len(formattedMembers),
				"members":       formattedMembers,
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *ListGuildMembersTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("list_guild_members", "List all members in a Discord server (guild)")
}

// formatMember formats a single member for the response
func (t *ListGuildMembersTool) formatMember(member *discordgo.Member) map[string]interface{} {
	return map[string]interface{}{
		"id":          member.User.ID,
		"username":    member.User.Username,
		"discriminator": member.User.Discriminator,
		"nick":        member.Nick,
		"roles":       member.Roles,
		"joined_at":   member.JoinedAt,
		"deaf":        member.Deaf,
		"mute":        member.Mute,
	}
}

// formatError creates a standardized error response
func (t *ListGuildMembersTool) formatError(message string, err error) types.CallToolResult {
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
