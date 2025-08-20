package logging

import (
	"io"
	"log/slog"
	"os"
)

func InitializeLogging(lvl slog.Level, sink io.Writer) {
	var addSource = false
	if lvl == slog.LevelDebug {
		addSource = true
	}
	options := slog.HandlerOptions{
		AddSource: addSource,
		Level:     lvl,
	}
	if sink == nil {
		sink = os.Stderr
	}

	logger := slog.New(slog.NewJSONHandler(sink, &options))
	slog.SetDefault(logger)
}
