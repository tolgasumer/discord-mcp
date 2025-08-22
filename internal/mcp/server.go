package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	"discord-mcp/internal/config"
	"discord-mcp/internal/discord"
	"discord-mcp/pkg/types"
)

// Server implements the MCP server
type Server struct {
	config      *config.Config
	logger      *logrus.Logger
	discord     *discord.Client
	tools       map[string]ToolHandler
	initialized bool
	mutex       sync.RWMutex
}

// ToolHandler defines the interface for tool handlers
type ToolHandler interface {
	Execute(params types.CallToolParams) (types.CallToolResult, error)
	GetDefinition() types.Tool
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Config, logger *logrus.Logger, discordClient *discord.Client) *Server {
	return &Server{
		config:  cfg,
		logger:  logger,
		discord: discordClient,
		tools:   make(map[string]ToolHandler),
	}
}

// RegisterTool registers a tool handler
func (s *Server) RegisterTool(handler ToolHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	tool := handler.GetDefinition()
	s.tools[tool.Name] = handler
	s.logger.Debugf("Registered tool: %s", tool.Name)
}

// Start starts the MCP server
func (s *Server) Start() error {
	s.logger.Info("Starting MCP server...")

	// Connect to Discord
	if err := s.discord.Connect(); err != nil {
		return fmt.Errorf("failed to connect to Discord: %w", err)
	}

	// Start handling stdin/stdout communication
	return s.handleCommunication(os.Stdin, os.Stdout)
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	s.logger.Info("Stopping MCP server...")
	
	if err := s.discord.Disconnect(); err != nil {
		s.logger.Warnf("Error disconnecting from Discord: %v", err)
	}
	
	return nil
}

// handleCommunication handles JSON-RPC communication over stdin/stdout
func (s *Server) handleCommunication(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		s.logger.Debugf("Received: %s", line)

		response := s.processMessage(line)
		if response != nil {
			responseJSON, err := json.Marshal(response)
			if err != nil {
				s.logger.Errorf("Failed to marshal response: %v", err)
				continue
			}

			s.logger.Debugf("Sending: %s", string(responseJSON))
			
			if _, err := fmt.Fprintln(output, string(responseJSON)); err != nil {
				s.logger.Errorf("Failed to write response: %v", err)
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// processMessage processes a single JSON-RPC message
func (s *Server) processMessage(message string) *types.Response {
	var req types.Request
	if err := json.Unmarshal([]byte(message), &req); err != nil {
		return &types.Response{
			JSONRPC: types.JSONRPCVersion,
			Error: &types.Error{
				Code:    types.ParseError,
				Message: "Parse error",
				Data:    err.Error(),
			},
		}
	}

	// Handle the request based on method
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "initialized":
		return s.handleInitialized(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolCall(req)
	case "ping":
		return s.handlePing(req)
	default:
		return &types.Response{
			JSONRPC: types.JSONRPCVersion,
			ID:      req.ID,
			Error: &types.Error{
				Code:    types.MethodNotFound,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req types.Request) *types.Response {
	var params types.InitializeParams
	if req.Params != nil {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &types.Response{
				JSONRPC: types.JSONRPCVersion,
				ID:      req.ID,
				Error: &types.Error{
					Code:    types.InvalidParams,
					Message: "Invalid parameters",
					Data:    err.Error(),
				},
			}
		}
	}

	s.logger.WithFields(logrus.Fields{
		"client_name":    params.ClientInfo.Name,
		"client_version": params.ClientInfo.Version,
	}).Info("Client initializing")

	result := types.InitializeResult{
		ProtocolVersion: types.ProtocolVersion,
		Capabilities: types.ServerCapabilities{
			Tools: &types.ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: types.ServerInfo{
			Name:    s.config.MCP.ServerName,
			Version: s.config.MCP.Version,
		},
	}

	return &types.Response{
		JSONRPC: types.JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
	}
}

// handleInitialized handles the initialized notification
func (s *Server) handleInitialized(req types.Request) *types.Response {
	s.mutex.Lock()
	s.initialized = true
	s.mutex.Unlock()
	
	s.logger.Info("Client initialized successfully")
	return nil // Notifications don't get responses
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req types.Request) *types.Response {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return &types.Response{
			JSONRPC: types.JSONRPCVersion,
			ID:      req.ID,
			Error: &types.Error{
				Code:    types.InvalidRequest,
				Message: "Server not initialized",
			},
		}
	}

	var tools []types.Tool
	for _, handler := range s.tools {
		tools = append(tools, handler.GetDefinition())
	}

	result := types.ToolsListResult{
		Tools: tools,
	}

	return &types.Response{
		JSONRPC: types.JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
	}
}

// handleToolCall handles the tools/call request
func (s *Server) handleToolCall(req types.Request) *types.Response {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return &types.Response{
			JSONRPC: types.JSONRPCVersion,
			ID:      req.ID,
			Error: &types.Error{
				Code:    types.InvalidRequest,
				Message: "Server not initialized",
			},
		}
	}

	var params types.CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &types.Response{
			JSONRPC: types.JSONRPCVersion,
			ID:      req.ID,
			Error: &types.Error{
				Code:    types.InvalidParams,
				Message: "Invalid parameters",
				Data:    err.Error(),
			},
		}
	}

	handler, exists := s.tools[params.Name]
	if !exists {
		return &types.Response{
			JSONRPC: types.JSONRPCVersion,
			ID:      req.ID,
			Error: &types.Error{
				Code:    types.MethodNotFound,
				Message: fmt.Sprintf("Tool not found: %s", params.Name),
			},
		}
	}

	s.logger.Debugf("Executing tool: %s", params.Name)
	result, err := handler.Execute(params)
	if err != nil {
		return &types.Response{
			JSONRPC: types.JSONRPCVersion,
			ID:      req.ID,
			Error: &types.Error{
				Code:    types.InternalError,
				Message: fmt.Sprintf("Tool execution failed: %v", err),
			},
		}
	}

	return &types.Response{
		JSONRPC: types.JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
	}
}

// handlePing handles ping requests
func (s *Server) handlePing(req types.Request) *types.Response {
	return &types.Response{
		JSONRPC: types.JSONRPCVersion,
		ID:      req.ID,
		Result:  map[string]string{"status": "pong"},
	}
}
