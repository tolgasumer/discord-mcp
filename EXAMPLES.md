# Real-World Usage Examples

This document provides practical, real-world examples of how the `discord-mcp` server can be used to power different kinds of applications. These scenarios are designed to illustrate the two primary interaction patterns: on-demand tool use and event-driven automation.

---

## Example 1: Powering a Conversational AI Assistant

In this scenario, a project manager uses a chatbot (like Claude, ChatGPT, or a custom Llama-based agent) to get a quick summary of recent activity from their team's Discord server. The chatbot uses `discord-mcp` as its toolset to interact with Discord.

**The Goal:** Get a summary of the last 3 announcements without leaving the chatbot interface.

### Step 1: The User's Prompt

The project manager types a natural language request to the AI assistant:

> **User:** Hey, can you check our `#announcements` channel and give me a summary of the last 3 posts?

### Step 2: The AI's Thought Process

The AI model receives the prompt and breaks it down into a series of actions. It understands it needs to interact with Discord, so it prepares to use the tools provided by `discord-mcp`.

1.  "The user wants messages from a channel named `#announcements`. I don't have the channel's ID, which I need for other tools. I must find it first."
2.  "I will call the `list_channels` tool to get a list of all available channels and find the one named `announcements`."
3.  "Once I have the ID, I will call the `get_channel_messages` tool, passing the channel ID and a `limit` of 3."
4.  "After I receive the message content, I will use my language capabilities to summarize them and present the result to the user."

### Step 3: MCP Interaction (Tool Calls)

The AI executes its plan by making JSON-RPC calls to the `discord-mcp` server.

**First, it calls `list_channels`:**
```json
// --> AI calls the server to find the channel ID
{
  "jsonrpc": "2.0",
  "id": 101,
  "method": "tools/call",
  "params": {
    "name": "list_channels",
    "arguments": {
      "guild_id": "123456789012345678"
    }
  }
}
```

**The server responds with a list of channels. The AI parses this to find the correct ID.**
```json
// <-- Server returns a list of channels
{
  "jsonrpc": "2.0",
  "id": 101,
  "result": {
    "channels": [
      { "id": "987654321098765432", "name": "general", "type": 0 },
      { "id": "876543210987654321", "name": "announcements", "type": 0 },
      { "id": "765432109876543210", "name": "dev-team", "type": 0 }
    ]
  }
}
```

**Now, with the ID `876543210987654321`, the AI calls `get_channel_messages`:**
```json
// --> AI calls the server to get the messages
{
  "jsonrpc": "2.0",
  "id": 102,
  "method": "tools/call",
  "params": {
    "name": "get_channel_messages",
    "arguments": {
      "channel_id": "876543210987654321",
      "limit": 3
    }
  }
}
```

### Step 4: The Final Response

The AI receives the content of the last three messages. It then generates a concise, natural language summary and presents it to the user.

> **Chatbot:** Of course! Here is a summary of the last three posts in `#announcements`:
> 1.  **Q3 Roadmap Published:** The product team has released the final roadmap for the third quarter.
> 2.  **API Downtime Notice:** There will be scheduled maintenance for the main API this Friday at 10 PM PST.
> 3.  **New Team Member:** Everyone welcome our new designer, Jane Doe!

This example shows how `discord-mcp` acts as the hands and eyes of a conversational AI, allowing it to perform complex tasks in external services.

---

## Example 2: Automated Moderation Bot

In this scenario, a community manager wants to build a simple, programmatic bot to improve the onboarding experience for new members. The bot will listen for new users joining and automatically assign them a role and send a welcome message.

**The Goal:** When a user joins the server, assign them the "New Member" role and send them a welcome DM.

### Step 1: The Trigger (Event Notification)

The `discord-mcp` server is running with the `discord/guildMemberAdded` event enabled in its configuration. When a new user joins the Discord server, it immediately sends a JSON-RPC notification to its connected client (the moderation bot).

```json
// <-- MCP server sends an unsolicited notification to the moderation bot
{
  "jsonrpc": "2.0",
  "method": "discord/guildMemberAdded",
  "params": {
    "guild_id": "123456789012345678",
    "user": {
      "id": "555666777888999000",
      "username": "new-user-jane"
    }
  }
}
```

### Step 2: The Bot's Programmatic Logic

The moderation bot is a simple application continuously listening for input from the `discord-mcp` server. Upon receiving the notification, its code is triggered.

Here is a Python pseudo-code example of the bot's logic:

```python
def handle_notification(notification):
    if notification.get("method") == "discord/guildMemberAdded":
        # Extract user info from the notification
        user_id = notification["params"]["user"]["id"]
        guild_id = notification["params"]["guild_id"]
        print(f"New user joined: {user_id}")

        # Action 1: Assign the 'New Member' role
        assign_role_payload = {
            "name": "assign_role",
            "arguments": {
                "guild_id": guild_id,
                "user_id": user_id,
                "role_name": "New Member"
            }
        }
        mcp_server.call_tool(assign_role_payload)

        # Action 2: Send a welcome message (not shown for brevity)
        # ... logic to call send_message to the user's DM ...

# Main loop of the bot
while True:
    incoming_message = mcp_server.read_line()
    if is_notification(incoming_message):
        handle_notification(incoming_message)
```

### Step 3: MCP Interaction (Tool Calls)

Based on the logic above, the moderation bot sends a new `tools/call` request back to the `discord-mcp` server to assign the role.

```json
// --> Moderation bot calls the server to assign a role
{
  "jsonrpc": "2.0",
  "id": 201,
  "method": "tools/call",
  "params": {
    "name": "assign_role",
    "arguments": {
      "guild_id": "123456789012345678",
      "user_id": "555666777888999000",
      "role_name": "New Member"
    }
  }
}
```

### Step 4: The Result

The `discord-mcp` server executes the request, and the new user `new-user-jane` now has the "New Member" role in Discord, all within milliseconds of joining.

This example shows how the event-streaming feature can be used to create powerful, proactive automations that respond to activity on Discord.
