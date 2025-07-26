package pkg

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

const (
	ansiReset          = "\033[0m"
	ansiFaint          = "\033[2m"
	ansiResetFaint     = "\033[22m"
	ansiBrightRed      = "\033[91m"
	ansiBrightGreen    = "\033[92m"
	ansiBrightYellow   = "\033[93m"
	ansiBrightRedFaint = "\033[91;2m"
	ansiBrightBlue     = "\033[94m"
)

func InitLogger(logLevel string) {
	var log slog.Level
	if logLevel == "info" {
		log = slog.LevelInfo
	} else {
		log = slog.LevelDebug
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      log,
			TimeFormat: time.Kitchen,
			NoColor:    false,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey {
					level, ok := a.Value.Any().(slog.Level)
					if !ok {
						return a
					}
					switch level {
					case slog.LevelError:
						a.Value = slog.StringValue(ansiBrightRed + "ERROR" + ansiReset)
					case slog.LevelWarn:
						a.Value = slog.StringValue(ansiBrightYellow + "WARN" + ansiReset)
					case slog.LevelInfo:
						a.Value = slog.StringValue(ansiBrightGreen + "INFO" + ansiReset)
					case slog.LevelDebug:
						a.Value = slog.StringValue(ansiBrightBlue + "DEBUG" + ansiReset)
					default:
						a.Value = slog.StringValue("UNKNOWN")
					}
				}
				return a
			},
		}),
	))
}

func Perf(msg string, start time.Time) {
	slog.Info(msg, "duration", time.Since(start))
}
