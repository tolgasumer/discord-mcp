# Discord MCP Server

A Model Context Protocol (MCP) server that provides Discord functionality to AI assistants and applications.

## Features

### Phase 1 (Completed)
- ✅ MCP server foundation with JSON-RPC 2.0 protocol
- ✅ Discord API integration using DiscordGo
- ✅ Configuration management with YAML and environment variables
- ✅ Rate limiting and connection management
- ✅ Basic health check tool (`ping`)
- ✅ Comprehensive logging and error handling
- ✅ Graceful shutdown handling

### Phase 2 (Completed)
- ✅ MCP server foundation with JSON-RPC 2.0 protocol
- ✅ Discord API integration using DiscordGo
- ✅ Configuration management with YAML and environment variables
- ✅ Rate limiting and connection management
- ✅ Basic health check tool (`ping`)
- ✅ Comprehensive logging and error handling
- ✅ Graceful shutdown handling

### Phase 3 (Current)
- ✅ Message operations (send, read, edit, delete)
- ✅ Channel management (list, create, info)
- ✅ Role management (list, get, create, delete, assign, unassign)
- ✅ Server information (guilds, members)

### Planned Features (Future Phases)
- Advanced Discord features (embeds, reactions, threads)

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
5. Under "Privileged Gateway Intents", enable:
   - Message Content Intent (if reading messages)
   - Server Members Intent (if accessing member info)

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

The server communicates via JSON-RPC 2.0 over stdin/stdout, following the MCP specification.

#### Available Tools

1. **ping** - Health check and bot status
   ```json
   {
     "jsonrpc": "2.0",
     "id": 1,
     "method": "tools/call",
     "params": {
       "name": "ping",
       "arguments": {}
     }
   }
   ```

#### Example MCP Session

```json
// Initialize
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "clientInfo": {"name": "test-client", "version": "1.0.0"}, "capabilities": {}}}

// List tools
{"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}}

// Call ping tool
{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "ping", "arguments": {}}}
```

## Architecture

```
discord-mcp/
├── cmd/discord-mcp/      # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── discord/         # Discord API client wrapper
│   ├── handlers/        # MCP tool handlers
│   └── mcp/            # MCP server implementation
├── pkg/types/          # Shared types and interfaces
├── config.yaml.example # Example configuration
├── go.mod             # Go module definition
└── README.md          # This file
```

### Key Components

- **MCP Server**: Handles JSON-RPC 2.0 protocol, tool registration, and client communication
- **Discord Client**: Wraps DiscordGo with rate limiting, error handling, and connection management
- **Tool Handlers**: Implement specific Discord operations as MCP tools
- **Configuration**: YAML-based config with environment variable overrides

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

1. Create a new handler in `internal/handlers/`
2. Implement the `ToolHandler` interface:
   ```go
   type ToolHandler interface {
       Execute(params types.CallToolParams) (types.CallToolResult, error)
       GetDefinition() types.Tool
   }
   ```
3. Register the tool in `main.go`

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

- **Phase 2**: Core Discord operations (send messages, get channels)
- **Phase 3**: Advanced features (embeds, reactions, role management)
- **Phase 4**: Production hardening and performance optimization
