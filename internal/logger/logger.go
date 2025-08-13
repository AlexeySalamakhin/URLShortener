package logger

import (
	"go.uber.org/zap"
)

// Log — глобальный логгер приложения. Инициализируется в Initialize.
var Log *zap.Logger = zap.NewNop()

// Initialize настраивает глобальный логгер по заданному уровню логирования.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}
