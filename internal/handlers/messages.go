package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"discord-mcp/internal/discord"
	"discord-mcp/internal/permissions"
	"discord-mcp/internal/validation"
	"discord-mcp/pkg/types"
)

// MessageHandler handles Discord message operations
type MessageHandler struct {
	discord     *discord.Client
	permissions *permissions.Checker
	validator   *validation.Validator
	logger      *logrus.Logger
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(discordClient *discord.Client, permChecker *permissions.Checker, validator *validation.Validator, logger *logrus.Logger) *MessageHandler {
	return &MessageHandler{
		discord:     discordClient,
		permissions: permChecker,
		validator:   validator,
		logger:      logger,
	}
}

// SendMessageTool implements the send_message MCP tool
type SendMessageTool struct {
	handler *MessageHandler
}

// NewSendMessageTool creates a new send message tool
func NewSendMessageTool(handler *MessageHandler) *SendMessageTool {
	return &SendMessageTool{handler: handler}
}

// Execute executes the send_message tool
func (t *SendMessageTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("send_message", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	channelID := params.Arguments["channel_id"].(string)
	content := params.Arguments["content"].(string)

	// Optional parameters
	tts := false
	if ttsVal, ok := params.Arguments["tts"]; ok {
		tts = ttsVal.(bool)
	}

	var replyTo string
	if replyVal, ok := params.Arguments["reply_to"]; ok {
		replyTo = replyVal.(string)
	}

	var embeds []*discordgo.MessageEmbed
	if embedsVal, ok := params.Arguments["embeds"]; ok {
		embedsSlice, ok := embedsVal.([]interface{})
		if !ok {
			return validation.FormatValidationError(fmt.Errorf("embeds must be an array")), nil
		}

		embeds = make([]*discordgo.MessageEmbed, len(embedsSlice))
		for i, embedData := range embedsSlice {
			embed, err := parseEmbed(embedData)
			if err != nil {
				return validation.FormatValidationError(fmt.Errorf("invalid embed at index %d: %w", i, err)), nil
			}
			embeds[i] = embed
		}
	}

	// Validate permissions
	extraData := map[string]interface{}{
		"tts": tts,
	}
	if err := t.handler.permissions.ValidateMessageOperation("send_message", channelID, extraData); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Prepare message data
	msgData := &discordgo.MessageSend{
		Content: content,
		TTS:     tts,
		Embeds:  embeds,
	}

	// Add reply reference if specified
	if replyTo != "" {
		msgData.Reference = &discordgo.MessageReference{
			MessageID: replyTo,
			ChannelID: channelID,
		}
	}

	// Send the message
	message, err := t.handler.discord.Session().ChannelMessageSendComplex(channelID, msgData)
	if err != nil {
		return t.formatError("Failed to send message", err), nil
	}

	// Format success response
	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("✅ Message sent successfully to <#%s>", channelID),
			Data: map[string]interface{}{
				"message_id":  message.ID,
				"channel_id":  channelID,
				"content":     message.Content,
				"timestamp":   message.Timestamp.Format(time.RFC3339),
				"tts":         message.TTS,
				"embed_count": len(message.Embeds),
				"has_reply":   replyTo != "",
				"message_url": fmt.Sprintf("https://discord.com/channels/%s/%s/%s", message.GuildID, channelID, message.ID),
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *SendMessageTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("send_message", "Send a message to a Discord channel with support for embeds, replies, and TTS")
}

// formatError creates a standardized error response
func (t *SendMessageTool) formatError(message string, err error) types.CallToolResult {
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

// GetChannelMessagesTool implements the get_channel_messages MCP tool
type GetChannelMessagesTool struct {
	handler *MessageHandler
}

// NewGetChannelMessagesTool creates a new get channel messages tool
func NewGetChannelMessagesTool(handler *MessageHandler) *GetChannelMessagesTool {
	return &GetChannelMessagesTool{handler: handler}
}

// Execute executes the get_channel_messages tool
func (t *GetChannelMessagesTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("get_channel_messages", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	channelID := params.Arguments["channel_id"].(string)

	limit := 50
	if limitVal, ok := params.Arguments["limit"]; ok {
		if limitFloat, ok := limitVal.(float64); ok {
			limit = int(limitFloat)
		} else if limitInt, ok := limitVal.(int); ok {
			limit = limitInt
		}
	}

	if limit > 100 {
		limit = 100
	}

	var beforeID, afterID, aroundID string
	if beforeVal, ok := params.Arguments["before"].(string); ok {
		beforeID = beforeVal
	}
	if afterVal, ok := params.Arguments["after"].(string); ok {
		afterID = afterVal
	}
	if aroundVal, ok := params.Arguments["around"].(string); ok {
		aroundID = aroundVal
	}

	// Validate permissions
	if err := t.handler.permissions.ValidateMessageOperation("get_messages", channelID, nil); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get messages from Discord
	messages, err := t.handler.discord.Session().ChannelMessages(channelID, limit, beforeID, afterID, aroundID)
	if err != nil {
		return t.formatError("Failed to get channel messages", err), nil
	}

	// Format messages for response
	formattedMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		formattedMessages[i] = t.formatMessage(msg)
	}

	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("📨 Retrieved %d messages from <#%s>", len(messages), channelID),
			Data: map[string]interface{}{
				"channel_id":    channelID,
				"message_count": len(messages),
				"messages":      formattedMessages,
				"query": map[string]interface{}{
					"limit":  limit,
					"before": beforeID,
					"after":  afterID,
					"around": aroundID,
				},
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *GetChannelMessagesTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("get_channel_messages", "Retrieve message history from a Discord channel with pagination support")
}

// formatMessage converts a Discord message to a structured format
func (t *GetChannelMessagesTool) formatMessage(msg *discordgo.Message) map[string]interface{} {
	// Format attachments
	attachments := make([]map[string]interface{}, len(msg.Attachments))
	for i, att := range msg.Attachments {
		attachments[i] = map[string]interface{}{
			"id":       att.ID,
			"filename": att.Filename,
			"size":     att.Size,
			"url":      att.URL,
			"width":    att.Width,
			"height":   att.Height,
		}
	}

	// Format embeds
	embeds := make([]map[string]interface{}, len(msg.Embeds))
	for i, embed := range msg.Embeds {
		embedData := map[string]interface{}{
			"title":       embed.Title,
			"description": embed.Description,
			"color":       embed.Color,
			"url":         embed.URL,
		}

		if embed.Thumbnail != nil {
			embedData["thumbnail"] = map[string]interface{}{
				"url": embed.Thumbnail.URL,
			}
		}

		if embed.Image != nil {
			embedData["image"] = map[string]interface{}{
				"url": embed.Image.URL,
			}
		}

		if len(embed.Fields) > 0 {
			fields := make([]map[string]interface{}, len(embed.Fields))
			for j, field := range embed.Fields {
				fields[j] = map[string]interface{}{
					"name":   field.Name,
					"value":  field.Value,
					"inline": field.Inline,
				}
			}
			embedData["fields"] = fields
		}

		embeds[i] = embedData
	}

	// Format reactions
	reactions := make([]map[string]interface{}, len(msg.Reactions))
	for i, reaction := range msg.Reactions {
		reactions[i] = map[string]interface{}{
			"emoji": map[string]interface{}{
				"name": reaction.Emoji.Name,
				"id":   reaction.Emoji.ID,
			},
			"count": reaction.Count,
			"me":    reaction.Me,
		}
	}

	return map[string]interface{}{
		"id":      msg.ID,
		"content": msg.Content,
		"author": map[string]interface{}{
			"id":            msg.Author.ID,
			"username":      msg.Author.Username,
			"discriminator": msg.Author.Discriminator,
			"avatar":        msg.Author.Avatar,
			"bot":           msg.Author.Bot,
		},
		"timestamp":        msg.Timestamp.Format(time.RFC3339),
		"edited":           msg.EditedTimestamp != nil,
		"tts":              msg.TTS,
		"mention_everyone": msg.MentionEveryone,
		"mentions":         t.formatMentions(msg.Mentions),
		"attachments":      attachments,
		"embeds":           embeds,
		"reactions":        reactions,
		"pinned":           msg.Pinned,
		"type":             int(msg.Type),
		"flags":            int(msg.Flags),
		"message_url":      fmt.Sprintf("https://discord.com/channels/%s/%s/%s", msg.GuildID, msg.ChannelID, msg.ID),
	}
}

// formatMentions formats user mentions
func (t *GetChannelMessagesTool) formatMentions(mentions []*discordgo.User) []map[string]interface{} {
	formatted := make([]map[string]interface{}, len(mentions))
	for i, user := range mentions {
		formatted[i] = map[string]interface{}{
			"id":            user.ID,
			"username":      user.Username,
			"discriminator": user.Discriminator,
			"avatar":        user.Avatar,
			"bot":           user.Bot,
		}
	}
	return formatted
}

// formatError creates a standardized error response
func (t *GetChannelMessagesTool) formatError(message string, err error) types.CallToolResult {
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

// EditMessageTool implements the edit_message MCP tool
type EditMessageTool struct {
	handler *MessageHandler
}

// NewEditMessageTool creates a new edit message tool
func NewEditMessageTool(handler *MessageHandler) *EditMessageTool {
	return &EditMessageTool{handler: handler}
}

// Execute executes the edit_message tool
func (t *EditMessageTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("edit_message", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	channelID := params.Arguments["channel_id"].(string)
	messageID := params.Arguments["message_id"].(string)

	var newContent string
	if contentVal, ok := params.Arguments["content"]; ok {
		newContent = contentVal.(string)
	}

	var newEmbeds []*discordgo.MessageEmbed
	if embedsVal, ok := params.Arguments["embeds"]; ok {
		embedsSlice, ok := embedsVal.([]interface{})
		if !ok {
			return validation.FormatValidationError(fmt.Errorf("embeds must be an array")), nil
		}

		newEmbeds = make([]*discordgo.MessageEmbed, len(embedsSlice))
		for i, embedData := range embedsSlice {
			embed, err := parseEmbed(embedData)
			if err != nil {
				return validation.FormatValidationError(fmt.Errorf("invalid embed at index %d: %w", i, err)), nil
			}
			newEmbeds[i] = embed
		}
	}

	// Validate permissions
	extraData := map[string]interface{}{
		"message_id": messageID,
	}
	if err := t.handler.permissions.ValidateMessageOperation("edit_message", channelID, extraData); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Prepare message edit data
	msgEdit := &discordgo.MessageEdit{
		Content: &newContent,
		ID:      messageID,
		Channel: channelID,
	}

	if newEmbeds != nil {
		msgEdit.Embeds = &newEmbeds
	}

	// Edit the message
	message, err := t.handler.discord.Session().ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return t.formatError("Failed to edit message", err), nil
	}

	// Format success response
	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("✏️ Message edited successfully in <#%s>", channelID),
			Data: map[string]interface{}{
				"message_id":       message.ID,
				"channel_id":       channelID,
				"new_content":      message.Content,
				"edited_timestamp": message.EditedTimestamp.Format(time.RFC3339),
				"embed_count":      len(message.Embeds),
				"message_url":      fmt.Sprintf("https://discord.com/channels/%s/%s/%s", message.GuildID, channelID, message.ID),
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *EditMessageTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("edit_message", "Edit a Discord message's content or embeds")
}

// formatError creates a standardized error response
func (t *EditMessageTool) formatError(message string, err error) types.CallToolResult {
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

// DeleteMessageTool implements the delete_message MCP tool
type DeleteMessageTool struct {
	handler *MessageHandler
}

// NewDeleteMessageTool creates a new delete message tool
func NewDeleteMessageTool(handler *MessageHandler) *DeleteMessageTool {
	return &DeleteMessageTool{handler: handler}
}

// Execute executes the delete_message tool
func (t *DeleteMessageTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("delete_message", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	channelID := params.Arguments["channel_id"].(string)
	messageID := params.Arguments["message_id"].(string)

	var reason string
	if reasonVal, ok := params.Arguments["reason"]; ok {
		reason = reasonVal.(string)
	}

	// Validate permissions
	extraData := map[string]interface{}{
		"message_id": messageID,
	}
	if err := t.handler.permissions.ValidateMessageOperation("delete_message", channelID, extraData); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Get message info before deletion (for logging)
	message, err := t.handler.discord.Session().ChannelMessage(channelID, messageID)
	if err != nil {
		return t.formatError("Failed to get message info before deletion", err), nil
	}

	// Delete the message
	err = t.handler.discord.Session().ChannelMessageDelete(channelID, messageID)
	if err != nil {
		return t.formatError("Failed to delete message", err), nil
	}

	// Log the deletion if reason provided
	if reason != "" {
		t.handler.logger.Infof("Deleted message %s in channel %s. Reason: %s", messageID, channelID, reason)
	}

	// Format success response
	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("🗑️ Message deleted successfully from <#%s>", channelID),
			Data: map[string]interface{}{
				"deleted_message_id": messageID,
				"channel_id":         channelID,
				"deleted_content":    message.Content,
				"author_id":          message.Author.ID,
				"author_username":    message.Author.Username,
				"deletion_reason":    reason,
				"deleted_at":         time.Now().Format(time.RFC3339),
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *DeleteMessageTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("delete_message", "Delete a Discord message")
}

// formatError creates a standardized error response
func (t *DeleteMessageTool) formatError(message string, err error) types.CallToolResult {
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

// AddReactionTool implements the add_reaction MCP tool
type AddReactionTool struct {
	handler *MessageHandler
}

// NewAddReactionTool creates a new add reaction tool
func NewAddReactionTool(handler *MessageHandler) *AddReactionTool {
	return &AddReactionTool{handler: handler}
}

// Execute executes the add_reaction tool
func (t *AddReactionTool) Execute(params types.CallToolParams) (types.CallToolResult, error) {
	// Validate parameters
	if err := t.handler.validator.ValidateToolParams("add_reaction", params.Arguments); err != nil {
		return validation.FormatValidationError(err), nil
	}

	// Extract parameters
	channelID := params.Arguments["channel_id"].(string)
	messageID := params.Arguments["message_id"].(string)
	emoji := params.Arguments["emoji"].(string)

	// Validate permissions
	extraData := map[string]interface{}{
		"emoji": emoji,
	}
	if err := t.handler.permissions.ValidateMessageOperation("add_reaction", channelID, extraData); err != nil {
		if permErr, ok := err.(*permissions.PermissionError); ok {
			return permissions.FormatPermissionError(permErr), nil
		}
		return t.formatError("Permission check failed", err), nil
	}

	// Validate and format emoji
	formattedEmoji := t.formatEmoji(emoji)
	if formattedEmoji == "" {
		return validation.FormatValidationError(fmt.Errorf("invalid emoji format: %s", emoji)), nil
	}

	// Add the reaction
	err := t.handler.discord.Session().MessageReactionAdd(channelID, messageID, formattedEmoji)
	if err != nil {
		return t.formatError("Failed to add reaction", err), nil
	}

	// Format success response
	return types.CallToolResult{
		Content: []types.Content{{
			Type: "text",
			Text: fmt.Sprintf("👍 Added reaction %s to message in <#%s>", emoji, channelID),
			Data: map[string]interface{}{
				"message_id":      messageID,
				"channel_id":      channelID,
				"emoji":           emoji,
				"formatted_emoji": formattedEmoji,
				"is_custom_emoji": t.isCustomEmoji(emoji),
				"added_at":        time.Now().Format(time.RFC3339),
				"message_url":     fmt.Sprintf("https://discord.com/channels/%s/%s/%s", "@me", channelID, messageID), // Guild ID is not available in this context, so we use @me to link to the channel.
			},
		}},
	}, nil
}

// GetDefinition returns the tool definition
func (t *AddReactionTool) GetDefinition() types.Tool {
	return validation.GetToolDefinition("add_reaction", "Add an emoji reaction to a Discord message")
}

// formatEmoji ensures the emoji is in the correct format for Discord API
func (t *AddReactionTool) formatEmoji(emoji string) string {
	// If it's already a custom emoji format <:name:id> or <a:name:id>, extract the name:id part
	if t.isCustomEmoji(emoji) {
		// Remove < and > and the initial : or a:
		if strings.HasPrefix(emoji, "<a:") {
			return emoji[3 : len(emoji)-1] // Remove <a: and >
		} else if strings.HasPrefix(emoji, "<:") {
			return emoji[2 : len(emoji)-1] // Remove <: and >
		}
	}

	// For Unicode emojis, return as-is
	return emoji
}

// parseEmbed converts interface{} to discordgo.MessageEmbed
func parseEmbed(embedData interface{}) (*discordgo.MessageEmbed, error) {
	embedMap, ok := embedData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("embed must be an object")
	}

	embed := &discordgo.MessageEmbed{}

	// Title
	if title, ok := embedMap["title"].(string); ok {
		embed.Title = title
	}

	// Description
	if description, ok := embedMap["description"].(string); ok {
		embed.Description = description
	}

	// Color
	if color, ok := embedMap["color"]; ok {
		if colorInt, ok := color.(int); ok {
			embed.Color = colorInt
		} else if colorFloat, ok := color.(float64); ok {
			embed.Color = int(colorFloat)
		}
	}

	// URL
	if url, ok := embedMap["url"].(string); ok {
		embed.URL = url
	}

	// Thumbnail
	if thumbnail, ok := embedMap["thumbnail"].(map[string]interface{}); ok {
		if thumbURL, ok := thumbnail["url"].(string); ok {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: thumbURL}
		}
	}

	// Image
	if image, ok := embedMap["image"].(map[string]interface{}); ok {
		if imgURL, ok := image["url"].(string); ok {
			embed.Image = &discordgo.MessageEmbedImage{URL: imgURL}
		}
	}

	// Fields
	if fields, ok := embedMap["fields"].([]interface{}); ok {
		embed.Fields = make([]*discordgo.MessageEmbedField, len(fields))
		for i, fieldData := range fields {
			fieldMap, ok := fieldData.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("field at index %d must be an object", i)
			}

			field := &discordgo.MessageEmbedField{}
			if name, ok := fieldMap["name"].(string); ok {
				field.Name = name
			}
			if value, ok := fieldMap["value"].(string); ok {
				field.Value = value
			}
			if inline, ok := fieldMap["inline"].(bool); ok {
				field.Inline = inline
			}

			embed.Fields[i] = field
		}
	}

	return embed, nil
}

// isCustomEmoji checks if an emoji is a custom Discord emoji
func (t *AddReactionTool) isCustomEmoji(emoji string) bool {
	return len(emoji) > 2 && emoji[0] == '<' && emoji[len(emoji)-1] == '>' && (strings.HasPrefix(emoji, "<:") || strings.HasPrefix(emoji, "<a:"))
}

// formatError creates a standardized error response
func (t *AddReactionTool) formatError(message string, err error) types.CallToolResult {
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
