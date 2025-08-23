# Discord MCP Server

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/your-repo/discord-mcp)

A Model Context Protocol (MCP) server that provides Discord functionality to AI assistants and applications.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Architecture](#architecture)
- [Development](#development)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)
- [License](#license)
- [Contributing](#contributing)
- [Roadmap](#roadmap)

## Features

The server exposes a set of tools that can be called by an MCP client. These tools provide a way to interact with the Discord API in a structured manner.

### General

- `ping`: Checks the health of the server and the connection to Discord.

### Guilds

- `get_guild_info`: Get information about a specific Discord server (guild).
- `list_guild_members`: List all members in a Discord server (guild).

### Channels

- `list_channels`: List channels in a Discord server (guild).
- `get_channel_info`: Get information about a specific Discord channel.

### Messages

- `send_message`: Sends a message to a Discord channel with support for embeds, replies, and TTS.
- `get_channel_messages`: Retrieves message history from a Discord channel with pagination support.
- `edit_message`: Edits a Discord message's content or embeds.
- `delete_message`: Deletes a Discord message.
- `add_reaction`: Adds an emoji reaction to a Discord message.

### Roles

- `list_roles`: Lists all roles in a Discord server (guild).
- `get_role_info`: Get information about a specific Discord role.
- `create_role`: Creates a new role in a Discord server (guild).
- `delete_role`: Deletes a role in a Discord server (guild).
- `assign_role`: Assigns a role to a user in a Discord server (guild).
- `unassign_role`: Unassigns a role from a user in a Discord server (guild).

### Event Streaming (Notifications)

Beyond the tool-based interaction, the server can stream real-time events from Discord directly to the MCP client. This is achieved through JSON-RPC notifications, allowing for proactive and responsive applications.

When a configured Discord event occurs, the server sends a notification like this:

```json
{
  "jsonrpc": "2.0",
  "method": "discord/messageCreated",
  "params": {
    "guild_id": "123456789012345678",
    "channel_id": "234567890123456789",
    "message_id": "345678901234567890",
    "author_id": "456789012345678901",
    "content": "Hello, world!"
  }
}
```

#### Supported Events
- `discord/messageCreated`: A new message is sent in a channel.
- `discord/guildMemberAdded`: A new user joins the guild.
- `discord/messageReactionAdded`: A reaction is added to a message.

Event streaming can be enabled and filtered in `config.yaml`.

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Discord bot token
- Discord application with appropriate permissions

### Installation

1. **Clone and build**:
   ```bash
   git clone <repository-url>
   cd discord-mcp
   go mod tidy
   go build -o discord-mcp ./cmd/discord-mcp
   ```

2. **Create configuration**:
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your Discord bot token
   ```

3. **Run the server**:
   ```bash
   ./discord-mcp
   ```

   Or with environment variable:
   ```bash
   DISCORD_TOKEN=your_bot_token_here ./discord-mcp
   ```

### Discord Bot Setup

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to "Bot" section and create a bot
4. Copy the bot token
5. Under "Privileged Gateway Intents", enable the intents required for your desired tools (see Permissions Guide below).

#### Bot Permissions Guide

To use all available tools, your bot will need the appropriate permissions in your Discord server's roles settings, as well as the correct Gateway Intents enabled in the Developer Portal.

| Tool Category | Required Permission(s) | Required Intent(s) |
|---------------|--------------------------------|--------------------|
| **Guilds** | `View Server As Member` | `Server Members` |
| **Channels** | `View Channels` | - |
| **Messages** | `Send Messages`, `Read Message History`, `Manage Messages` (for edit/delete), `Add Reactions` | `Message Content` |
| **Roles** | `Manage Roles` | - |

*Note: Granting the `Administrator` permission will cover all permission requirements, but is not recommended for production bots.*

## Configuration

### Configuration File (`config.yaml`)

```yaml
discord:
  token: "YOUR_BOT_TOKEN_HERE"
  guild_id: ""                    # Optional default guild
  allowed_guilds: []              # Restrict to specific guilds
  max_message_length: 2000        # Discord's limit
  rate_limit_per_minute: 30       # Rate limiting

mcp:
  server_name: "discord-mcp"
  version: "1.0.0"

events:
  enabled: true                   # Master switch for all events
  allowed_events:                 # List of events to stream
    - "discord/messageCreated"
    - "discord/guildMemberAdded"
    - "discord/messageReactionAdded"

server:
  log_level: "info"               # debug, info, warn, error
  debug: false
```

### Environment Variables
- `DISCORD_TOKEN` - Discord bot token (overrides config)
- `DISCORD_GUILD_ID` - Default guild ID
- `LOG_LEVEL` - Log level

## Usage

### Command Line Options

```bash
./discord-mcp [options]

Options:
  -config string
        Path to configuration file (default "config.yaml")
  -log-level string
        Log level (debug, info, warn, error)
  -version
        Show version and exit
```

### MCP Client Integration

The server supports two primary modes of interaction with an MCP client: synchronous tool calls (request/response) and asynchronous event streaming (notifications).

#### Tool-Based Interaction (Request/Response)

For clients that need to execute specific actions on demand, the server acts as a standard JSON-RPC tool provider. The client sends a `tools/call` request, and the server responds with the result.

**Example Session:**
```json
// --> Client sends initialize request
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "clientInfo": {"name": "test-client", "version": "1.0.0"}}}

// <-- Server responds with its capabilities
{"jsonrpc": "2.0", "id": 1, "result": {"protocolVersion": "1.0.0", "serverInfo": {"name": "discord-mcp"}}}

// --> Client asks to send a message
{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "send_message", "arguments": {"channel_id": "1234567890", "content": "Hello from my AI assistant!"}}}

// <-- Server confirms the message was sent and returns its details
{"jsonrpc": "2.0", "id": 2, "result": {"message_id": "9876543210", "content": "Hello from my AI assistant!", "author_id": "bot-id-here"}}
```

#### Event-Based Interaction (Notifications)

For clients that need to react to events in real-time, the server can be configured to stream Discord events as JSON-RPC notifications. This is optional and can be enabled via the `events` section in `config.yaml`.

When a subscribed event occurs on Discord, the server will send an unsolicited notification to the client. The client should be prepared to listen for these messages at any time after initialization.

**Example Notification:**

Here is an example of a notification the client might receive when a new user joins the server. Note that it has no `id` field, as it is not a response to a request.

```json
// <-- Server sends a notification about a new member
{
  "jsonrpc": "2.0",
  "method": "discord/guildMemberAdded",
  "params": {
    "guild_id": "123456789012345678",
    "user": {
      "id": "543210987654321098",
      "username": "new-user"
    }
  }
}
```

## Architecture

The server is designed with a clean separation of concerns. The following diagram illustrates the high-level data flow between the client, the server's components, and the Discord API.

```
+-----------------+      +------------------+      +------------------+
|                 |      |                  |      |                  |
|   MCP Client    |----->|    MCP Server    |<---->|  Discord Client  |
| (e.g., AI Agent)|      |  (stdin/stdout)  |      |  (discordgo)     |
|                 |      |                  |      |                  |
+-------+---------+      +--------+---------+      +---------+--------+
        ^                         |                        |
        | (Notifications)         | (Tool Calls)           | (API Events)
        |                         |                        |
+-------+---------+               |                        v
|                 |               v                +------------------+
| Notification Svc|      +--------+---------+      |                  |
|                 |      |  Tool Handlers  |      |   Discord API    |
+-----------------+      +------------------+      |                  |
                                                   +------------------+
```

```
discord-mcp/
├── cmd/discord-mcp/      # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── discord/         # Discord API client wrapper
│   ├── handlers/        # MCP tool handlers
│   ├── mcp/             # MCP server implementation
│   └── notifications/   # Event notification service
├── pkg/types/          # Shared types and interfaces
├── config.yaml.example # Example configuration
├── go.mod             # Go module definition
└── README.md          # This file
```

### Key Components

- **MCP Server**: Handles JSON-RPC 2.0 protocol, tool registration, and client communication.
- **Discord Client**: Wraps DiscordGo with rate limiting, error handling, and connection management.
- **Notification Service**: Formats and sends asynchronous JSON-RPC notifications for Discord events.
- **Tool Handlers**: Implement specific Discord operations as MCP tools.
- **Configuration**: YAML-based config with environment variable overrides.

## Development

### Building

```bash
go build -o discord-mcp ./cmd/discord-mcp
```

### Testing

```bash
go test ./...
```

### Adding New Tools

1. Create a new handler in `internal/handlers/`.
2. Implement the `ToolHandler` interface:
   ```go
   type ToolHandler interface {
       Execute(params types.CallToolParams) (types.CallToolResult, error)
       GetDefinition() types.Tool
   }
   ```
3. Register the tool in `cmd/discord-mcp/main.go`.

### Adding New Events

1.  **Find the event in `discordgo`**: Identify the event you want to handle from the `bwmarrin/discordgo` library (e.g., `*discordgo.ChannelCreate`).

2.  **Create a handler in `internal/discord/dispatcher.go`**: Add a new `Handle...` method to the `EventDispatcher`.

    ```go
    func (d *EventDispatcher) HandleChannelCreate(s *discordgo.Session, e *discordgo.ChannelCreate) {
        // 1. Check if event is enabled
        if !d.config.Enabled || !d.isEventAllowed("discord/channelCreated") {
            return
        }

        // 2. Construct parameters
        params := map[string]interface{}{
            "guild_id":   e.GuildID,
            "channel_id": e.ID,
            "name":       e.Name,
            "type":       e.Type,
        }

        // 3. Create and send notification
        notification := d.createNotification("discord/channelCreated", params)
        if err := d.notificationSvc.Send(notification); err != nil {
            d.logger.Errorf("Failed to send channelCreated notification: %v", err)
        }
    }
    ```

3.  **Register the handler in `internal/discord/client.go`**: In the `SetupEventHandlers` function, add the new handler to the session.

    ```go
    c.session.AddHandler(c.dispatcher.HandleChannelCreate)
    ```

4.  **Update configuration**: Add the new event name (`discord/channelCreated`) to the `allowed_events` list in `config.yaml.example` so users know it's available.

### Debug Mode

Run with debug logging:
```bash
./discord-mcp -log-level debug
```

## Security Considerations

- Bot tokens are sensitive - never commit them to version control
- Use environment variables or secure configuration management
- Restrict guild access using `allowed_guilds` configuration
- Rate limiting is implemented but respect Discord's API limits
- Validate all inputs in tool handlers

## Troubleshooting

### Common Issues

1. **"Discord token is required"**
   - Set token in config.yaml or DISCORD_TOKEN environment variable

2. **"Failed to connect to Discord"**
   - Verify bot token is valid
   - Check bot permissions in Discord server
   - Ensure bot is added to the intended server

3. **Rate limit exceeded**
   - Reduce `rate_limit_per_minute` in configuration
   - Check if multiple instances are running

4. **Tool execution fails**
   - Check bot permissions for the specific Discord operation
   - Verify guild IDs and channel IDs are correct
   - Review logs with `-log-level debug`

### Logging

The server provides comprehensive logging:
- Connection status and events
- Tool execution details
- Rate limiting information
- Error details with stack traces (in debug mode)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Roadmap

This roadmap is prioritized by business impact.

### High Impact

- **Event Streaming:** Stream Discord events (e.g., new messages, user joins) to the MCP client for real-time, proactive AI assistants.
- **Interaction Support:** Full support for slash commands, buttons, select menus, and modals.
- **Thread Management:** Tools to create, delete, and manage threads.
- **User Management:** Tools to kick, ban, and get info on users for moderation.

### Medium Impact

- **Voice Channel Management:** Join, leave, and manage voice channels, and play audio.
- **Advanced Permissions:** Granular permission system for tool execution.
- **Auto-Moderation:** Manage Discord's auto-moderation rules.
- **Reaction Management:** Add, remove, and list reactions on messages.

### Low Impact

- **Performance and Optimization:** Benchmark and optimize the server for high-traffic use.
- **Granular Get Tools:** More specific `get` tools (e.g., list banned users).
- **Scheduled Events:** Tools to manage scheduled guild events.
