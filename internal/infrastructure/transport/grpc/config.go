package transportgrpc

import (
	"time"

	"github.com/kvoloboi/telemetry/internal/application/common"
)

type SenderConfig struct {
	MaxReconnectAttempts    int
	Backoff                 common.Backoff
	CloseOnServerDisconnect bool
	Buffer                  int
}

func defaultSenderConfig() SenderConfig {
	return SenderConfig{
		MaxReconnectAttempts:    5,
		Backoff:                 common.NewBackoff(100*time.Millisecond, 5*time.Second),
		CloseOnServerDisconnect: false,
		Buffer:                  100,
	}
}
