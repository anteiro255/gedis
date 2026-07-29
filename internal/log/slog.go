package log

import (
	"log/slog"
	"os"

	"github.com/anteiro255/gedis/internal/config"
)

func InitLogger(cfg *config.Config) {
	var logLevel slog.Level
	switch cfg.Verbosity() {
	case 0:
		logLevel = slog.LevelError
	case 1:
		logLevel = slog.LevelWarn
	case 2:
		logLevel = slog.LevelInfo
	case 3:
		logLevel = slog.LevelDebug
	default:
		slog.Error("Unexpected verbosity level. Internal error")
	}

	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{
					AddSource: true,
					Level:     logLevel,
				},
			),
		),
	)
}
