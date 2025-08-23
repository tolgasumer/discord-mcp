package notifications

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"discord-mcp/pkg/types"
	"github.com/sirupsen/logrus"
)

// Service handles sending JSON-RPC notifications.
type Service struct {
	writer io.Writer
	logger *logrus.Logger
	mutex  sync.Mutex
}

// NewService creates a new notification service.
func NewService(writer io.Writer, logger *logrus.Logger) *Service {
	return &Service{
		writer: writer,
		logger: logger,
	}
}

// Send marshals and sends a notification to the client.
func (s *Service) Send(notification *types.Notification) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.writer == nil {
		return fmt.Errorf("notification writer not configured")
	}

	notification.JSONRPC = types.JSONRPCVersion
	responseJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	s.logger.Debugf("Sending notification: %s", string(responseJSON))

	if _, err := fmt.Fprintln(s.writer, string(responseJSON)); err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return nil
}
