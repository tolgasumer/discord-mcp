package validation

import "discord-mcp/pkg/types"

// ToolSchemas defines JSON schemas for all Discord MCP tools
var ToolSchemas = map[string]interface{}{
	"send_message": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"minLength":   1,
				"description": "Discord channel ID (snowflake)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"minLength":   1,
				"maxLength":   2000,
				"description": "Message content (Discord markdown supported)",
			},
			"tts": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Whether message should be read aloud using TTS",
			},
			"reply_to": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Message ID to reply to",
			},
			"embeds": map[string]interface{}{
				"type":        "array",
				"maxItems":    10,
				"description": "Array of embed objects",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":      "string",
							"maxLength": 256,
						},
						"description": map[string]interface{}{
							"type":      "string",
							"maxLength": 4096,
						},
						"color": map[string]interface{}{
							"type":    "integer",
							"minimum": 0,
							"maximum": 16777215,
						},
						"url": map[string]interface{}{
							"type":   "string",
							"format": "uri",
						},
						"thumbnail": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"url": map[string]interface{}{
									"type":   "string",
									"format": "uri",
								},
							},
							"required": []string{"url"},
						},
						"image": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"url": map[string]interface{}{
									"type":   "string",
									"format": "uri",
								},
							},
							"required": []string{"url"},
						},
						"fields": map[string]interface{}{
							"type":     "array",
							"maxItems": 25,
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type":      "string",
										"maxLength": 256,
									},
									"value": map[string]interface{}{
										"type":      "string",
										"maxLength": 1024,
									},
									"inline": map[string]interface{}{
										"type":    "boolean",
										"default": false,
									},
								},
								"required": []string{"name", "value"},
							},
						},
					},
				},
			},
		},
		"required": []string{"channel_id", "content"},
	},

	"get_channel_messages": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Discord channel ID (snowflake)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"maximum":     100,
				"default":     50,
				"description": "Number of messages to retrieve (1-100)",
			},
			"before": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Get messages before this message ID",
			},
			"after": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Get messages after this message ID",
			},
			"around": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Get messages around this message ID",
			},
		},
		"required": []string{"channel_id"},
		"not": map[string]interface{}{
			"anyOf": []map[string]interface{}{
				{"allOf": []map[string]interface{}{
					{"required": []string{"before"}},
					{"required": []string{"after"}},
				}},
				{"allOf": []map[string]interface{}{
					{"required": []string{"before"}},
					{"required": []string{"around"}},
				}},
				{"allOf": []map[string]interface{}{
					{"required": []string{"after"}},
					{"required": []string{"around"}},
				}},
			},
		},
	},

	"edit_message": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Discord channel ID (snowflake)",
			},
			"message_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Message ID to edit",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"maxLength":   2000,
				"description": "New message content",
			},
			"embeds": map[string]interface{}{
				"type":        "array",
				"maxItems":    10,
				"description": "New embed objects",
			},
		},
		"required": []string{"channel_id", "message_id"},
		"anyOf": []map[string]interface{}{
			{"required": []string{"content"}},
			{"required": []string{"embeds"}},
		},
	},

	"delete_message": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Discord channel ID (snowflake)",
			},
			"message_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Message ID to delete",
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"maxLength":   512,
				"description": "Reason for deletion (appears in audit log)",
			},
		},
		"required": []string{"channel_id", "message_id"},
	},

	"add_reaction": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Discord channel ID (snowflake)",
			},
			"message_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Message ID to react to",
			},
			"emoji": map[string]interface{}{
				"type":        "string",
				"description": "Emoji to add (Unicode emoji or custom emoji format)",
			},
		},
		"required": []string{"channel_id", "message_id", "emoji"},
	},

	"list_channels": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID to list channels from",
			},
			"type_filter": map[string]interface{}{
				"type":        "array",
				"description": "Filter channels by type",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{"text", "voice", "category", "announcement", "stage", "forum", "media"},
				},
				"uniqueItems": true,
			},
			"include_permissions": map[string]interface{}{
				"type":        "boolean",
				"default":     false,
				"description": "Include bot permissions for each channel",
			},
		},
		"required": []string{"guild_id"},
	},

	"get_channel_info": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"channel_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Discord channel ID (snowflake)",
			},
			"include_permissions": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include bot permissions for this channel",
			},
		},
		"required": []string{"channel_id"},
	},

	"get_guild_info": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID",
			},
			"include_counts": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Include member and channel counts",
			},
		},
		"required": []string{"guild_id"},
	},

	"list_guild_members": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID",
			},
		},
		"required": []string{"guild_id"},
	},

	"get_role_info": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID",
			},
			"role_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Role ID",
			},
		},
		"required": []string{"guild_id", "role_id"},
	},

	"create_role": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the new role",
			},
		},
		"required": []string{"guild_id", "name"},
	},

	"delete_role": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID",
			},
			"role_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Role ID to delete",
			},
		},
		"required": []string{"guild_id", "role_id"},
	},

	"assign_role": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID",
			},
			"role_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Role ID to assign",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "User ID to assign the role to",
			},
		},
		"required": []string{"guild_id", "role_id", "user_id"},
	},

	"unassign_role": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID",
			},
			"role_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Role ID to unassign",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "User ID to unassign the role from",
			},
		},
		"required": []string{"guild_id", "role_id", "user_id"},
	},

	"list_roles": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"guild_id": map[string]interface{}{
				"type":        "string",
				"pattern":     "^[0-9]+$",
				"description": "Guild (server) ID to list roles from",
			},
		},
		"required": []string{"guild_id"},
	},
}

// GetToolSchema returns the JSON schema for a specific tool
func GetToolSchema(toolName string) (interface{}, bool) {
	schema, exists := ToolSchemas[toolName]
	return schema, exists
}

// GetToolDefinition returns a types.Tool with the schema for a given tool
func GetToolDefinition(toolName, description string) types.Tool {
	schema, exists := GetToolSchema(toolName)
	if !exists {
		// Return a basic schema if not found
		schema = map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		}
	}

	return types.Tool{
		Name:        toolName,
		Description: description,
		InputSchema: schema,
	}
}
