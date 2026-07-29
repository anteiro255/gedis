package log

import (
	"log/slog"
	"os"
)

func InitLogger() {

	var logLevel = slog.LevelInfo
	if os.Getenv("GEDIS_VERBOSE") == "1" {
		logLevel = slog.LevelDebug
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
