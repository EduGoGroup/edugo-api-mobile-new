package mock

import (
	"context"

	"github.com/EduGoGroup/edugo-shared/audit"
)

// MockAuditLogger is a mock implementation of audit.AuditLogger for tests.
type MockAuditLogger struct {
	LogFn   func(ctx context.Context, event audit.AuditEvent) error
	Called  bool
	Events  []audit.AuditEvent
}

func (m *MockAuditLogger) Log(ctx context.Context, event audit.AuditEvent) error {
	m.Called = true
	m.Events = append(m.Events, event)
	if m.LogFn != nil {
		return m.LogFn(ctx, event)
	}
	return nil
}

// NewNoopAuditLogger returns a no-op audit logger for tests.
func NewNoopAuditLogger() audit.AuditLogger {
	return audit.NewNoopAuditLogger()
}
