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

// RoleHandler handles Discord role operations
type RoleHandler struct {
	discord     *discord.Client
	permissions *permissions.Checker
	validator   *validation.Validator
	logger      *logrus.Logger
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(discordClient *discord.Client, permChecker *permissions.Checker, validator *validation.Validator, logger *logrus.Logger) *RoleHandler {
	return &RoleHandler{
		discord:     discordClient,
		permissions: permChecker,
		validator:   validator,
		logger:      logger,
	}
}

// ListRolesTool implements the list_roles MCP tool
type ListRolesTool struct {
	handler *RoleHandler
}

// NewListRolesTool creates a new list roles tool
func NewListRolesTool(handler *RoleHandler) *ListRolesTool {
	return &ListRolesTool{handler: handler}
}

// Execute executes the list_roles tool
func (t *ListRolesTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("list_roles", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanManageRoles(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get roles from Discord
	roles, err := t.handler.discord.Session().GuildRoles(guildID)
	if err != nil {
		return t.formatError("Failed to list roles", err), nil
	}

	// Format roles for response
	formattedRoles := make([]map[string]interface{}, len(roles))
	for i, role := range roles {
		formattedRoles[i] = t.formatRole(role)
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Found %d roles in guild %s", len(formattedRoles), guildID),
			Data: map[string]interface{}{
				"guild_id":   guildID,
				"role_count": len(formattedRoles),
				"roles":      formattedRoles,
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *ListRolesTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("list_roles", "List all roles in a Discord server (guild)")
}

// formatRole formats a single role for the response
func (t *ListRolesTool) formatRole(role *discordgo.Role) map[string]interface{} {
	return map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"color":       role.Color,
		"hoist":       role.Hoist,
		"position":    role.Position,
		"permissions": role.Permissions,
		"managed":     role.Managed,
		"mentionable": role.Mentionable,
	}
}

// formatError creates a standardized error response
func (t *ListRolesTool) formatError(message string, err error) types.CallToolResult {
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

// GetRoleInfoTool implements the get_role_info MCP tool
type GetRoleInfoTool struct {
	handler *RoleHandler
}

// NewGetRoleInfoTool creates a new get role info tool
func NewGetRoleInfoTool(handler *RoleHandler) *GetRoleInfoTool {
	return &GetRoleInfoTool{handler: handler}
}

// Execute executes the get_role_info tool
func (t *GetRoleInfoTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("get_role_info", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)
	roleID := params.Arguments["role_id"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanManageRoles(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get role from Discord
	role, err := t.handler.discord.Session().State.Role(guildID, roleID)
	if err != nil {
		return t.formatError("Failed to get role info", err), nil
	}

	// Format role for response
	formattedRole := t.formatRole(role)

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Role: %s", role.Name),
			Data: formattedRole,
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *GetRoleInfoTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("get_role_info", "Get information about a specific Discord role")
}

// formatRole formats a single role for the response
func (t *GetRoleInfoTool) formatRole(role *discordgo.Role) map[string]interface{} {
	return map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"color":       role.Color,
		"hoist":       role.Hoist,
		"position":    role.Position,
		"permissions": role.Permissions,
		"managed":     role.Managed,
		"mentionable": role.Mentionable,
	}
}

// formatError creates a standardized error response
func (t *GetRoleInfoTool) formatError(message string, err error) types.CallToolResult {
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

// CreateRoleTool implements the create_role MCP tool
type CreateRoleTool struct {
	handler *RoleHandler
}

// NewCreateRoleTool creates a new create role tool
func NewCreateRoleTool(handler *RoleHandler) *CreateRoleTool {
	return &CreateRoleTool{handler: handler}
}

// Execute executes the create_role tool
func (t *CreateRoleTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("create_role", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)
	name := params.Arguments["name"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanManageRoles(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Create role
	role, err := t.handler.discord.Session().GuildRoleCreate(guildID, &discordgo.RoleParams{Name: name})
	if err != nil {
		return t.formatError("Failed to create role", err), nil
	}

	// Format role for response
	formattedRole := t.formatRole(role)

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Created role: %s", role.Name),
			Data: formattedRole,
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *CreateRoleTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("create_role", "Create a new role in a Discord server (guild)")
}

// formatRole formats a single role for the response
func (t *CreateRoleTool) formatRole(role *discordgo.Role) map[string]interface{} {
	return map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"color":       role.Color,
		"hoist":       role.Hoist,
		"position":    role.Position,
		"permissions": role.Permissions,
		"managed":     role.Managed,
		"mentionable": role.Mentionable,
	}
}

// formatError creates a standardized error response
func (t *CreateRoleTool) formatError(message string, err error) types.CallToolResult {
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

// DeleteRoleTool implements the delete_role MCP tool
type DeleteRoleTool struct {
	handler *RoleHandler
}

// NewDeleteRoleTool creates a new delete role tool
func NewDeleteRoleTool(handler *RoleHandler) *DeleteRoleTool {
	return &DeleteRoleTool{handler: handler}
}

// Execute executes the delete_role tool
func (t *DeleteRoleTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("delete_role", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)
	roleID := params.Arguments["role_id"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanManageRoles(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Delete role
	if err := t.handler.discord.Session().GuildRoleDelete(guildID, roleID); err != nil {
		return t.formatError("Failed to delete role", err), nil
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Deleted role with ID: %s", roleID),
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *DeleteRoleTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("delete_role", "Delete a role in a Discord server (guild)")
}

// formatError creates a standardized error response
func (t *DeleteRoleTool) formatError(message string, err error) types.CallToolResult {
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

// AssignRoleTool implements the assign_role MCP tool
type AssignRoleTool struct {
	handler *RoleHandler
}

// NewAssignRoleTool creates a new assign role tool
func NewAssignRoleTool(handler *RoleHandler) *AssignRoleTool {
	return &AssignRoleTool{handler: handler}
}

// Execute executes the assign_role tool
func (t *AssignRoleTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("assign_role", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)
	roleID := params.Arguments["role_id"].(string)
	userID := params.Arguments["user_id"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanManageRoles(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Assign role
	if err := t.handler.discord.Session().GuildMemberRoleAdd(guildID, userID, roleID); err != nil {
		return t.formatError("Failed to assign role", err), nil
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Assigned role %s to user %s", roleID, userID),
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *AssignRoleTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("assign_role", "Assign a role to a user in a Discord server (guild)")
}

// formatError creates a standardized error response
func (t *AssignRoleTool) formatError(message string, err error) types.CallToolResult {
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

// UnassignRoleTool implements the unassign_role MCP tool
type UnassignRoleTool struct {
	handler *RoleHandler
}

// NewUnassignRoleTool creates a new unassign role tool
func NewUnassignRoleTool(handler *RoleHandler) *UnassignRoleTool {
	return &UnassignRoleTool{handler: handler}
}

// Execute executes the unassign_role tool
func (t *UnassignRoleTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("unassign_role", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	guildID := params.Arguments["guild_id"].(string)
	roleID := params.Arguments["role_id"].(string)
	userID := params.Arguments["user_id"].(string)

	// Validate permissions
	if err := t.handler.permissions.CanManageRoles(guildID); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Unassign role
	if err := t.handler.discord.Session().GuildMemberRoleRemove(guildID, userID, roleID); err != nil {
		return t.formatError("Failed to unassign role", err), nil
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("Unassigned role %s from user %s", roleID, userID),
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *UnassignRoleTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("unassign_role", "Unassign a role from a user in a Discord server (guild)")
}

// formatError creates a standardized error response
func (t *UnassignRoleTool) formatError(message string, err error) types.CallToolResult {
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
