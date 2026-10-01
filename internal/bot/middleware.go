package bot

import (
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram"
	"go.uber.org/zap"
)

func GetFloodMiddleware(log *zap.Logger) []telegram.Middleware {
	waiter := floodwait.NewSimpleWaiter().WithMaxRetries(10)
	return []telegram.Middleware{
		waiter,
	}
}
