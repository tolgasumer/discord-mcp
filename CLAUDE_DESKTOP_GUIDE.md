# Using the MCP Server with Claude Desktop

This guide explains how to connect and use the Discord MCP (Model Context Protocol) Server with the Claude Desktop application.

## Introduction

The Discord MCP Server is a powerful bridge that connects a Large Language Model (LLM) like Claude to the Discord API. It is a Go application that you will compile and then configure Claude Desktop to launch.

This guide will walk you through the process of compiling the `discord-mcp` server and configuring the Claude Desktop application to use it.

## Prerequisites

* You have a Discord bot token. You can learn how to create a Discord bot and get a token from the [Discord Developer Portal](https://discord.com/developers/docs/intro).
* You have the Claude Desktop application installed.
* You have Go 1.21 or higher installed on your system.

## Step 1: Build the MCP Server

First, you need to compile the `discord-mcp` Go application to create an executable file.

1.  **Open a terminal or command prompt.**
2.  **Navigate to the root directory of the `discord-mcp` project.**
3.  **Build the application:**

    ```bash
    go build -o discord-mcp ./cmd/discord-mcp
    ```

This command will create an executable file named `discord-mcp` (or `discord-mcp.exe` on Windows) in the project's root directory.

## Step 2: Configure Claude Desktop

Now, you will use the built-in editor in Claude Desktop to configure it to launch your compiled `discord-mcp` server.

### 1. Open the Configuration Editor

1.  **Open Claude Desktop Settings:** Launch the Claude Desktop application and navigate to the settings section.
2.  **Go to Developer Settings:** In the settings menu, find and click on the **Developer** section.
3.  **Edit Configuration:** Click on the **Edit Config** button. This will open the `claude_desktop_config.json` file in your default text editor.

### 2. Add the MCP Server Configuration

Replace the content of the `claude_desktop_config.json` file with the following JSON structure:

```json
{
  "mcpServers": {
    "discord-mcp": {
      "command": "path/to/your/discord-mcp/executable",
      "args": [],
      "env": {
        "DISCORD_TOKEN": "YOUR_DISCORD_BOT_TOKEN_HERE"
      }
    }
  }
}
```

**Important:**

*   **`command`**: Replace `"path/to/your/discord-mcp/executable"` with the **absolute path** to the `discord-mcp` executable you created in Step 1.
    *   **Windows Example:** `"D:\\repo\\discord-mcp\\discord-mcp.exe"` (Note the double backslashes).
    *   **UNIX (Linux/macOS) Example:** `"/home/user/discord-mcp/discord-mcp"`.
*   **`args`**: This field can be left as an empty array `[]` if you are using the default `config.yaml` file name and location.
*   **`env`**: This is where you provide environment variables to the server. 
    *   Replace `"YOUR_DISCORD_BOT_TOKEN_HERE"` with your actual Discord bot token.


### 3. Save and Restart

After saving the `claude_desktop_config.json` file, restart the Claude Desktop application. Claude Desktop will now automatically start and manage the `discord-mcp` server process for you.

## Usage

Once Claude Desktop is configured, you can start using the `discord-mcp` tools in your conversations with Claude. For example, you could say:

> "Using the discord-mcp tool, send a message to the #general channel saying 'Hello from Claude!'"

Claude will then use the `send_message` tool from your `discord-mcp` server to send the message to the specified channel.

## Troubleshooting

*   **Claude doesn't see the tools:**
    *   Ensure you have correctly formatted the `claude_desktop_config.json` file. It should be valid JSON.
    *   Make sure you have restarted Claude Desktop after editing the file.
*   **Server fails to start:**
    *   Double-check that the path to your `discord-mcp` executable in the `command` field is correct and that it is an absolute path.
    *   Make sure the executable has the necessary permissions to run.
*   **Tool execution fails:**
    *   Check the `discord-mcp` server logs for any error messages. You may need to run the executable manually in a terminal to see the logs.
    *   Verify that your Discord bot has the necessary permissions for the action you are trying to perform.