package discord

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"discord-mcp/internal/config"
	"discord-mcp/internal/notifications"
)

// Client wraps the Discord session and provides higher-level operations
type Client struct {
	session    *discordgo.Session
	config     *config.Config
	logger     *logrus.Logger
	dispatcher *EventDispatcher

	// Connection state
	connected bool
	mutex     sync.RWMutex

	// Rate limiting
	rateLimiter *rateLimiter
}

// rateLimiter implements simple rate limiting
type rateLimiter struct {
	requests []time.Time
	maxReqs  int
	duration time.Duration
	mutex    sync.Mutex
}

// NewClient creates a new Discord client
func NewClient(cfg *config.Config, logger *logrus.Logger) (*Client, error) {
	// Create Discord session
	session, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Configure session
	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessageReactions

	client := &Client{
		session:     session,
		config:      cfg,
		logger:      logger,
		rateLimiter: newRateLimiter(cfg.Discord.RateLimitPerMinute, time.Minute),
	}

	return client, nil
}

// SetupEventHandlers sets up the event handlers for the Discord client
func (c *Client) SetupEventHandlers(notificationSvc *notifications.Service) {
	c.dispatcher = NewEventDispatcher(c.logger, notificationSvc, &c.config.Events)

	c.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		c.logger.WithFields(logrus.Fields{
			"username": r.User.Username,
			"id":       r.User.ID,
		}).Info("Discord bot is ready")
	})

	c.session.AddHandler(func(s *discordgo.Session, d *discordgo.Disconnect) {
		c.logger.Warn("Disconnected from Discord")
	})

	c.session.AddHandler(c.dispatcher.HandleMessageCreate)
	c.session.AddHandler(c.dispatcher.HandleGuildMemberAdd)
	c.session.AddHandler(c.dispatcher.HandleMessageReactionAdd)
}

// Connect connects to Discord
func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return nil
	}

	c.logger.Info("Connecting to Discord...")

	if err := c.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}

	c.connected = true
	c.logger.Info("Connected to Discord successfully")
	return nil
}

// Disconnect disconnects from Discord
func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	c.logger.Info("Disconnecting from Discord...")

	if err := c.session.Close(); err != nil {
		return fmt.Errorf("failed to close Discord connection: %w", err)
	}

	c.connected = false
	c.logger.Info("Disconnected from Discord")
	return nil
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.connected
}

// GetBotUser returns information about the bot user
func (c *Client) GetBotUser() (*discordgo.User, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Discord")
	}

	if !c.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	return c.session.State.User, nil
}

// GetGuild returns information about a guild
func (c *Client) GetGuild(guildID string) (*discordgo.Guild, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Discord")
	}

	if !c.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Check if guild is allowed
	if !c.isGuildAllowed(guildID) {
		return nil, fmt.Errorf("access to guild %s is not allowed", guildID)
	}

	guild, err := c.session.Guild(guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	return guild, nil
}

// GetChannels returns all channels in a guild
func (c *Client) GetChannels(guildID string) ([]*discordgo.Channel, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Discord")
	}

	if !c.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Check if guild is allowed
	if !c.isGuildAllowed(guildID) {
		return nil, fmt.Errorf("access to guild %s is not allowed", guildID)
	}

	channels, err := c.session.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	return channels, nil
}

// SendMessage sends a message to a channel
func (c *Client) SendMessage(channelID, content string) (*discordgo.Message, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Discord")
	}

	if !c.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Validate message length
	if len(content) > c.config.Discord.MaxMessageLength {
		return nil, fmt.Errorf("message exceeds maximum length of %d characters",
			c.config.Discord.MaxMessageLength)
	}

	message, err := c.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Debugf("Sent message to channel %s", channelID)
	return message, nil
}

// GetChannelMessages returns recent messages from a channel
func (c *Client) GetChannelMessages(channelID string, limit int) ([]*discordgo.Message, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Discord")
	}

	if !c.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Discord API limit is 100
	if limit > 100 {
		limit = 100
	}

	messages, err := c.session.ChannelMessages(channelID, limit, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get channel messages: %w", err)
	}

	return messages, nil
}

// Ping tests the connection to Discord
func (c *Client) Ping() error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to Discord")
	}

	// Try to get bot user as a simple connectivity test
	_, err := c.GetBotUser()
	return err
}

// setupEventHandlers sets up Discord event handlers
func (c *Client) setupEventHandlers() {
	c.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		c.logger.WithFields(logrus.Fields{
			"username": r.User.Username,
			"id":       r.User.ID,
		}).Info("Discord bot is ready")
	})

	c.session.AddHandler(func(s *discordgo.Session, d *discordgo.Disconnect) {
		c.logger.Warn("Disconnected from Discord")
	})

	// Register event dispatcher handlers
	c.session.AddHandler(c.dispatcher.HandleMessageCreate)
	c.session.AddHandler(c.dispatcher.HandleGuildMemberAdd)
	c.session.AddHandler(c.dispatcher.HandleMessageReactionAdd)
}

// Session returns the underlying DiscordGo session for advanced operations
func (c *Client) Session() *discordgo.Session {
	return c.session
}

// isGuildAllowed checks if the guild is in the allowed list (if configured)
func (c *Client) isGuildAllowed(guildID string) bool {
	// If no restrictions are configured, allow all guilds
	if len(c.config.Discord.AllowedGuilds) == 0 {
		return true
	}

	// Check if the guild is in the allowed list
	for _, allowedGuild := range c.config.Discord.AllowedGuilds {
		if allowedGuild == guildID {
			return true
		}
	}

	return false
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(maxReqs int, duration time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make([]time.Time, 0),
		maxReqs:  maxReqs,
		duration: duration,
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *rateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Remove old requests outside the time window
	cutoff := now.Add(-rl.duration)
	var validReqs []time.Time
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validReqs = append(validReqs, req)
		}
	}
	rl.requests = validReqs

	// Check if we can make a new request
	if len(rl.requests) >= rl.maxReqs {
		return false
	}

	// Add the new request
	rl.requests = append(rl.requests, now)
	return true
}
