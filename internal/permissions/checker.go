package permissions

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"discord-mcp/internal/discord"
	"discord-mcp/pkg/types"
)

// Checker validates Discord permissions before operations
type Checker struct {
	discord *discord.Client
	logger  *logrus.Logger
}

// NewChecker creates a new permission checker
func NewChecker(discordClient *discord.Client, logger *logrus.Logger) *Checker {
	return &Checker{
		discord: discordClient,
		logger:  logger,
	}
}

// PermissionError represents a permission-related error
type PermissionError struct {
	Operation   string `json:"operation"`
	Permission  string `json:"permission"`
	Resource    string `json:"resource"`
	Description string `json:"description"`
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("insufficient permissions for %s on %s: missing %s (%s)", 
		e.Operation, e.Resource, e.Permission, e.Description)
}

// NewPermissionError creates a new permission error
func NewPermissionError(operation, permission, resource, description string) *PermissionError {
	return &PermissionError{
		Operation:   operation,
		Permission:  permission,
		Resource:    resource,
		Description: description,
	}
}

// FormatPermissionError returns a properly formatted MCP tool result for permission errors
func FormatPermissionError(err *PermissionError) types.CallToolResult {
	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("🔒 Permission Error: %s", err.Description),
			Data: map[string]interface{}{
				"error_type":  "permission",
				"operation":   err.Operation,
				"permission":  err.Permission,
				"resource":    err.Resource,
				"description": err.Description,
			},
		}},
		IsError: true,
	}
}

// Channel Permission Methods

// CanSendMessages checks if the bot can send messages to a channel
func (c *Checker) CanSendMessages(channelID string) error {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionSendMessages == 0 {
		return NewPermissionError("send_message", "SEND_MESSAGES", 
			fmt.Sprintf("channel:%s", channelID), 
			"Bot cannot send messages to this channel")
	}

	return nil
}

// CanSendTTSMessages checks if the bot can send TTS messages to a channel
func (c *Checker) CanSendTTSMessages(channelID string) error {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionSendTTSMessages == 0 {
		return NewPermissionError("send_tts_message", "SEND_TTS_MESSAGES",
			fmt.Sprintf("channel:%s", channelID),
			"Bot cannot send TTS messages to this channel")
	}

	return nil
}

// CanReadMessageHistory checks if the bot can read message history in a channel
func (c *Checker) CanReadMessageHistory(channelID string) error {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionReadMessageHistory == 0 {
		return NewPermissionError("read_messages", "READ_MESSAGE_HISTORY",
			fmt.Sprintf("channel:%s", channelID),
			"Bot cannot read message history in this channel")
	}

	return nil
}

// CanManageMessages checks if the bot can manage (edit/delete) messages in a channel
func (c *Checker) CanManageMessages(channelID string) error {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionManageMessages == 0 {
		return NewPermissionError("manage_messages", "MANAGE_MESSAGES",
			fmt.Sprintf("channel:%s", channelID),
			"Bot cannot manage messages in this channel")
	}

	return nil
}

// CanAddReactions checks if the bot can add reactions to messages
func (c *Checker) CanAddReactions(channelID string) error {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionAddReactions == 0 {
		return NewPermissionError("add_reaction", "ADD_REACTIONS",
			fmt.Sprintf("channel:%s", channelID),
			"Bot cannot add reactions in this channel")
	}

	return nil
}

// CanUseExternalEmojis checks if the bot can use external emojis
func (c *Checker) CanUseExternalEmojis(channelID string) error {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionUseExternalEmojis == 0 {
		return NewPermissionError("use_external_emoji", "USE_EXTERNAL_EMOJIS",
			fmt.Sprintf("channel:%s", channelID),
			"Bot cannot use external emojis in this channel")
	}

	return nil
}

// CanViewChannel checks if the bot can view a channel
func (c *Checker) CanViewChannel(channelID string) error {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionViewChannel == 0 {
		return NewPermissionError("view_channel", "VIEW_CHANNEL",
			fmt.Sprintf("channel:%s", channelID),
			"Bot cannot view this channel")
	}

	return nil
}

// Guild Permission Methods

// CanViewGuild checks if the bot can view guild information
func (c *Checker) CanViewGuild(guildID string) error {
	// Check if the bot is in the guild
	guild, err := c.discord.GetGuild(guildID)
	if err != nil {
		return NewPermissionError("view_guild", "GUILD_ACCESS",
			fmt.Sprintf("guild:%s", guildID),
			"Bot does not have access to this guild")
	}

	if guild == nil {
		return NewPermissionError("view_guild", "GUILD_ACCESS",
			fmt.Sprintf("guild:%s", guildID),
			"Guild not found or bot not a member")
	}

	return nil
}

// CanManageRoles checks if the bot can manage roles in a guild
func (c *Checker) CanManageRoles(guildID string) error {
	permissions, err := c.getBotGuildPermissions(guildID)
	if err != nil {
		return err
	}

	if permissions&discordgo.PermissionManageRoles == 0 {
		return NewPermissionError("manage_roles", "MANAGE_ROLES",
			fmt.Sprintf("guild:%s", guildID),
			"Bot cannot manage roles in this guild")
	}

	return nil
}

// Message-specific Permission Methods

// CanEditMessage checks if the bot can edit a specific message
func (c *Checker) CanEditMessage(channelID, messageID string) error {
	// First check if we can manage messages in general
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	// Bots can always edit their own messages
	// They need MANAGE_MESSAGES to edit others' messages
	message, err := c.getMessageInfo(channelID, messageID)
	if err != nil {
		return err
	}

	botUser, err := c.discord.GetBotUser()
	if err != nil {
		return fmt.Errorf("failed to get bot user info: %w", err)
	}

	// If it's the bot's own message, just need send permissions
	if message.Author.ID == botUser.ID {
		if permissions&discordgo.PermissionSendMessages == 0 {
			return NewPermissionError("edit_own_message", "SEND_MESSAGES",
				fmt.Sprintf("message:%s", messageID),
				"Bot cannot edit its own messages without SEND_MESSAGES permission")
		}
		return nil
	}

	// For other users' messages, need MANAGE_MESSAGES
	if permissions&discordgo.PermissionManageMessages == 0 {
		return NewPermissionError("edit_others_message", "MANAGE_MESSAGES",
			fmt.Sprintf("message:%s", messageID),
			"Bot cannot edit other users' messages without MANAGE_MESSAGES permission")
	}

	return nil
}

// CanDeleteMessage checks if the bot can delete a specific message
func (c *Checker) CanDeleteMessage(channelID, messageID string) error {
	// Similar logic to edit message
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return err
	}

	message, err := c.getMessageInfo(channelID, messageID)
	if err != nil {
		return err
	}

	botUser, err := c.discord.GetBotUser()
	if err != nil {
		return fmt.Errorf("failed to get bot user info: %w", err)
	}

	// If it's the bot's own message, just need manage messages OR it's a recent message
	if message.Author.ID == botUser.ID {
		// Bots can delete their own messages
		return nil
	}

	// For other users' messages, need MANAGE_MESSAGES
	if permissions&discordgo.PermissionManageMessages == 0 {
		return NewPermissionError("delete_others_message", "MANAGE_MESSAGES",
			fmt.Sprintf("message:%s", messageID),
			"Bot cannot delete other users' messages without MANAGE_MESSAGES permission")
	}

	return nil
}

// Helper Methods

// getUserChannelPermissions gets the bot's permissions for a specific channel
func (c *Checker) getUserChannelPermissions(channelID string) (int64, error) {
	botUser, err := c.discord.GetBotUser()
	if err != nil {
		return 0, fmt.Errorf("failed to get bot user: %w", err)
	}

	// Get channel info to determine guild
	channel, err := c.getChannelInfo(channelID)
	if err != nil {
		return 0, err
	}

	if channel.GuildID == "" {
		// DM channel - bots have basic permissions in DMs
		return discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionAddReactions, nil
	}

	// Get user permissions in the channel
	permissions, err := c.discord.Session().UserChannelPermissions(botUser.ID, channelID)
	if err != nil {
		return 0, fmt.Errorf("failed to get channel permissions: %w", err)
	}

	return permissions, nil
}

// getChannelInfo gets basic channel information
func (c *Checker) getChannelInfo(channelID string) (*discordgo.Channel, error) {
	channel, err := c.discord.Session().Channel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}
	return channel, nil
}

// getMessageInfo gets basic message information
func (c *Checker) getMessageInfo(channelID, messageID string) (*discordgo.Message, error) {
	message, err := c.discord.Session().ChannelMessage(channelID, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message info: %w", err)
	}
	return message, nil
}

// getBotGuildPermissions gets the bot's permissions for a specific guild
func (c *Checker) getBotGuildPermissions(guildID string) (int64, error) {
	botUser, err := c.discord.GetBotUser()
	if err != nil {
		return 0, fmt.Errorf("failed to get bot user: %w", err)
	}

	_, err = c.discord.GetGuild(guildID)
	if err != nil {
		return 0, err
	}

	member, err := c.discord.Session().State.Member(guildID, botUser.ID)
	if err != nil {
		member, err = c.discord.Session().GuildMember(guildID, botUser.ID)
		if err != nil {
			return 0, fmt.Errorf("failed to get bot member info: %w", err)
		}
	}

	var permissions int64
	for _, roleID := range member.Roles {
		role, err := c.discord.Session().State.Role(guildID, roleID)
		if err != nil {
			return 0, fmt.Errorf("failed to get role info: %w", err)
		}
		permissions |= role.Permissions
	}

	return permissions, nil
}

// GetChannelPermissions returns a summary of bot permissions for a channel
func (c *Checker) GetChannelPermissions(channelID string) (map[string]bool, error) {
	permissions, err := c.getUserChannelPermissions(channelID)
	if err != nil {
		return nil, err
	}

	return map[string]bool{
		"view_channel":           permissions&discordgo.PermissionViewChannel != 0,
		"send_messages":          permissions&discordgo.PermissionSendMessages != 0,
		"send_tts_messages":      permissions&discordgo.PermissionSendTTSMessages != 0,
		"manage_messages":        permissions&discordgo.PermissionManageMessages != 0,
		"read_message_history":   permissions&discordgo.PermissionReadMessageHistory != 0,
		"add_reactions":          permissions&discordgo.PermissionAddReactions != 0,
		"use_external_emojis":    permissions&discordgo.PermissionUseExternalEmojis != 0,
		"attach_files":           permissions&discordgo.PermissionAttachFiles != 0,
		"embed_links":            permissions&discordgo.PermissionEmbedLinks != 0,
		"mention_everyone":       permissions&discordgo.PermissionMentionEveryone != 0,
	}, nil
}

// ValidateMessageOperation performs comprehensive permission checking for message operations
func (c *Checker) ValidateMessageOperation(operation, channelID string, extraData map[string]interface{}) error {
	c.logger.Debugf("Validating %s operation for channel %s", operation, channelID)

	// First, check if we can view the channel
	if err := c.CanViewChannel(channelID); err != nil {
		return err
	}

	switch operation {
	case "send_message":
		if err := c.CanSendMessages(channelID); err != nil {
			return err
		}
		
		// Check TTS if requested
		if tts, ok := extraData["tts"].(bool); ok && tts {
			if err := c.CanSendTTSMessages(channelID); err != nil {
				return err
			}
		}

	case "get_messages":
		if err := c.CanReadMessageHistory(channelID); err != nil {
			return err
		}

	case "edit_message":
		if messageID, ok := extraData["message_id"].(string); ok {
			if err := c.CanEditMessage(channelID, messageID); err != nil {
				return err
			}
		} else {
			return NewPermissionError("edit_message", "MESSAGE_ID_REQUIRED", channelID, "Message ID is required for edit operations")
		}

	case "delete_message":
		if messageID, ok := extraData["message_id"].(string); ok {
			if err := c.CanDeleteMessage(channelID, messageID); err != nil {
				return err
			}
		} else {
			return NewPermissionError("delete_message", "MESSAGE_ID_REQUIRED", channelID, "Message ID is required for delete operations")
		}

	case "add_reaction":
		if err := c.CanAddReactions(channelID); err != nil {
			return err
		}
		
		// Check external emoji if needed
		if emoji, ok := extraData["emoji"].(string); ok && c.isExternalEmoji(emoji) {
			if err := c.CanUseExternalEmojis(channelID); err != nil {
				return err
			}
		}

	default:
		c.logger.Warnf("Unknown operation for permission validation: %s", operation)
	}

	return nil
}

// isExternalEmoji checks if an emoji string represents an external (custom) emoji
func (c *Checker) isExternalEmoji(emoji string) bool {
	// Custom emojis have the format <:name:id> or <a:name:id> for animated
	return len(emoji) > 2 && emoji[0] == '<' && emoji[len(emoji)-1] == '>'
}
