package logutils

import (
	"fmt"

	"github.com/rs/zerolog"
)

const (
	colorRed = iota + 31
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorBlueLight
	_
	colorBold = 1
)

// ConsoleFormatLevel returns a custom colorizer for zerolog console level output
func ConsoleFormatLevel() zerolog.Formatter {
	return func(i any) string {
		if ll, ok := i.(string); ok {
			switch ll {
			case "trace":
				return "|"
			case "debug":
				return colorize("|", colorBlueLight)
			case "info":
				return colorize("|", colorGreen)
			case "warn":
				return colorize("|", colorYellow)
			case "error":
				return colorize("|", colorRed)
			case "fatal":
				return colorize(colorize("|", colorRed), colorBold)
			case "panic":
				return colorize("|", colorRed)
			default:
				return "|"
			}
		}
		return "|"
	}
}

// ConsoleFormatErrFieldName returns a custom formatter for error field name.
func ConsoleFormatErrFieldName() zerolog.Formatter {
	return func(i any) string {
		return fmt.Sprintf("%s=", i)
	}
}

// ConsoleFormatErrFieldValue returns a custom formatter for error value.
func ConsoleFormatErrFieldValue() zerolog.Formatter {
	return func(i any) string {
		return fmt.Sprintf("%s", i)
	}
}

// colorize returns the string s wrapped in ANSI code c
func colorize(s any, c int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

// wrap returns the string s wrapped in ANSI code c
func wrap(s string) string {
	return fmt.Sprintf("[%s]", s)
}
