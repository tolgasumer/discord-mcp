# Using the MCP Server with Claude Desktop

This guide explains how to connect and use the Discord MCP (Model Context Protocol) Server with the Claude Desktop application.

## Introduction

The Discord MCP Server is a bridge that connects a LLMs like Claude to the Discord API. It allows Claude to interact with Discord by calling predefined tools and receiving real-time events.

This guide will walk you through the process of configuring the Claude Desktop to use your running `discord-mcp` server.

## Prerequisites

Before you begin, ensure you have the following:

*   The MCP Server is installed and running. You should know the address and port it is running on (e.g., `http://localhost:8080`).
*   You have a Discord bot token. You can learn how to create a Discord bot and get a token from the [Discord Developer Portal](https://discord.com/developers/docs/intro).
*   You have the Claude Desktop application installed.

## Configuring the MCP Server

Make sure your `discord-mcp` server is configured with your Discord bot token in the `config.yaml` file as described in the main `README.md`.

## Configuring Claude Desktop

There are two ways to configure Claude Desktop to use your `discord-mcp` server. The recommended method is to use the built-in Extensions UI.

### Method 1: Using the Built-in Editor (Recommended)

The Claude Desktop app provides a built-in editor to manage your tool servers.

1.  **Open Claude Desktop Settings:** Launch the Claude Desktop application and navigate to the settings section.
2.  **Go to Developer Settings:** In the settings menu, find and click on the **Developer** section.
3.  **Edit Configuration:** Click on the **Edit Config** button. This will open the `claude_desktop_config.json` file in your default text editor.
4.  **Add the MCP Server:** Add or edit the `mcpServers` array in the JSON file as below.
```json
{
  "mcpServers": [
    {
      "name": "discord-mcp",
      "url": "http://localhost:8080",
      "tools": [
        "ping",
        "get_guild_info",
        "list_guild_members",
        "list_channels",
        "get_channel_info",
        "send_message",
        "get_channel_messages",
        "edit_message",
        "delete_message",
        "add_reaction",
        "list_roles",
        "get_role_info",
        "create_role",
        "delete_role",
        "assign_role",
        "unassign_role"
      ]
    }
  ]
}
```
5. **Save and Restart:** Save the file and restart Claude Desktop for the changes to take effect.

Once restarted, Claude Desktop will be able to communicate with your `discord-mcp` server.

## Usage

Once Claude Desktop is configured, you can start using the `discord-mcp` tools in your conversations with Claude. For example, you could say:

> "Using the discord-mcp tool, send a message to the #general channel saying 'Hello from Claude!'"

Claude will then use the `send_message` tool from your `discord-mcp` server to send the message to the specified channel.

## Troubleshooting
*   **Tool execution fails:**
    *   Check the `discord-mcp` server logs for any error messages.
    *   Ensure your `discord-mcp` server is running and accessible from your computer.
    *   Verify that your Discord bot has the necessary permissions for the action you are trying to perform.