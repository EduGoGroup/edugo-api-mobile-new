package messaging

import (
	"github.com/EduGoGroup/edugo-shared/messaging/rabbit"
)

// NewPublisher creates a RabbitMQ publisher from a connection.
func NewPublisher(conn *rabbit.Connection) rabbit.Publisher {
	return rabbit.NewPublisher(conn)
}
