package logutils

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	config "github.com/zs-dima/auth-service/cmd/config"
)

func SetupLogging(config *config.LogConfig) (*zerolog.Logger, *os.File) {
	configureConsoleWriter(config)

	logLevel, ok := logLevelMatches[config.Level]
	if !ok {
		logLevel = zerolog.InfoLevel
	}

	var logger zerolog.Logger
	var f *os.File
	if config.Human || isTerminalAttached() {
		zerolog.TimeFieldFormat = time.RFC3339Nano
		writer := configureConsoleWriter(config)
		logger = zerolog.New(writer).Level(logLevel).With().Timestamp().Logger()
	} else if config.File == "" {
		logger = zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().Logger()
	} else {
		var err error
		f, err = os.OpenFile(config.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Warn().Msgf("error opening log file: %v, falling back to console", err)
			logger = zerolog.New(os.Stdout).Level(logLevel).With().Timestamp().Logger()
		} else {
			logger = zerolog.New(f).Level(logLevel).With().Timestamp().Logger()
		}
	}

	zerolog.SetGlobalLevel(logLevel)

	return &logger, f
}

// ConfigureConsoleWriter returns a console writer with custom configuration.
func configureConsoleWriter(config *config.LogConfig) *zerolog.ConsoleWriter {
	return &zerolog.ConsoleWriter{
		Out:                 os.Stdout,
		TimeFormat:          "04:05.00",
		FormatErrFieldName:  ConsoleFormatErrFieldName(),
		FormatErrFieldValue: ConsoleFormatErrFieldValue(),
		FormatLevel:         ConsoleFormatLevel(),
	}
}

func isTerminalAttached() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) && runtime.GOOS != "windows"
}

var logLevelMatches = map[string]zerolog.Level{
	"NONE":  zerolog.NoLevel,
	"TRACE": zerolog.TraceLevel,
	"DEBUG": zerolog.DebugLevel,
	"INFO":  zerolog.InfoLevel,
	"WARN":  zerolog.WarnLevel,
	"ERROR": zerolog.ErrorLevel,
	"FATAL": zerolog.FatalLevel,
}

func InterceptorLogger(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
