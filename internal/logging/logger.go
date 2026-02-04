package logging

import (
	"log/slog"
	"github.com/lmittmann/tint"
	"os"
)

func Init() {
	w := os.Stdout
	logger := slog.New(tint.NewHandler(w, &tint.Options{
		Level:       slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}
