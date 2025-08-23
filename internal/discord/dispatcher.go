package discord

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"discord-mcp/internal/config"
	"discord-mcp/internal/notifications"
	"discord-mcp/pkg/types"
)

// EventDispatcher handles Discord events and dispatches them to the MCP client
type EventDispatcher struct {
	logger          *logrus.Logger
	notificationSvc *notifications.Service
	config          *config.EventsConfig
}

// NewEventDispatcher creates a new EventDispatcher
func NewEventDispatcher(logger *logrus.Logger, notificationSvc *notifications.Service, config *config.EventsConfig) *EventDispatcher {
	return &EventDispatcher{
		logger:          logger,
		notificationSvc: notificationSvc,
		config:          config,
	}
}

// HandleMessageCreate handles the MessageCreate event from Discord
func (d *EventDispatcher) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !d.config.Enabled || !d.isEventAllowed("discord/messageCreated") {
		return
	}
	d.logger.Debugf("Handling MessageCreate event for message ID: %s", m.ID)

	params := map[string]interface{}{
		"guild_id":   m.GuildID,
		"channel_id": m.ChannelID,
		"message_id": m.ID,
		"author_id":  m.Author.ID,
		"content":    m.Content,
	}

	if err := d.notificationSvc.Send(d.createNotification("discord/messageCreated", params)); err != nil {
		d.logger.Errorf("Failed to send messageCreated notification: %v", err)
	}
}

// HandleGuildMemberAdd handles the GuildMemberAdd event from Discord
func (d *EventDispatcher) HandleGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if !d.config.Enabled || !d.isEventAllowed("discord/guildMemberAdded") {
		return
	}
	d.logger.Debugf("Handling GuildMemberAdd event for user ID: %s", m.User.ID)

	params := map[string]interface{}{
		"guild_id": m.GuildID,
		"user": map[string]interface{}{
			"id":       m.User.ID,
			"username": m.User.Username,
		},
	}

	if err := d.notificationSvc.Send(d.createNotification("discord/guildMemberAdded", params)); err != nil {
		d.logger.Errorf("Failed to send guildMemberAdded notification: %v", err)
	}
}

// HandleMessageReactionAdd handles the MessageReactionAdd event from Discord
func (d *EventDispatcher) HandleMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !d.config.Enabled || !d.isEventAllowed("discord/messageReactionAdded") {
		return
	}
	d.logger.Debugf("Handling MessageReactionAdd event for message ID: %s", r.MessageID)

	params := map[string]interface{}{
		"guild_id":   r.GuildID,
		"channel_id": r.ChannelID,
		"message_id": r.MessageID,
		"user_id":    r.UserID,
		"emoji": map[string]interface{}{
			"id":   r.Emoji.ID,
			"name": r.Emoji.Name,
		},
	}

	if err := d.notificationSvc.Send(d.createNotification("discord/messageReactionAdded", params)); err != nil {
		d.logger.Errorf("Failed to send messageReactionAdded notification: %v", err)
	}
}

func (d *EventDispatcher) createNotification(method string, params map[string]interface{}) *types.Notification {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		// This should not happen with the defined structures
		panic(fmt.Sprintf("failed to marshal notification params: %v", err))
	}

	return &types.Notification{
		JSONRPC: types.JSONRPCVersion,
		Method:  method,
		Params:  paramsJSON,
	}
}

func (d *EventDispatcher) isEventAllowed(event string) bool {
	for _, allowedEvent := range d.config.AllowedEvents {
		if allowedEvent == event {
			return true
		}
	}
	return false
}
