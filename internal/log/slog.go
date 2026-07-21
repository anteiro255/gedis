package log

import (
	"log/slog"
	"os"
)

func InitLogger() {
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				os.Stdout,
				&slog.HandlerOptions{
					AddSource: true,
				},
			),
		),
	)
}
