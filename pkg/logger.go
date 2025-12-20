package pkg

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var Log *LoggerService

type LoggerService struct {
	Logger zerolog.Logger
	Env    string
}

func InitLogger(env string) (*LoggerService, error) {
	var output io.Writer

	switch env {
	case "DEVELOPMENT":
		file, err := os.OpenFile("dev.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}
		fileWriter := zerolog.ConsoleWriter{
			Out:        file,
			TimeFormat: "",
			FormatFieldName: func(i any) string {
				return fmt.Sprintf("%s=", i)
			},
			FormatFieldValue: func(i any) string {
				s := fmt.Sprintf("%v", i)
				if strings.ContainsAny(s, " \t\n\r") {
					return fmt.Sprintf("%q", s)
				}
				return s
			},
			FormatTimestamp: func(i any) string {
				t, err := time.Parse(time.RFC3339, i.(string))
				if err != nil {
					return fmt.Sprintf("time=%q", i) // Fallback if parsing fails
				}
				return fmt.Sprintf("time=%d", t.UnixMilli())
			},
			FormatLevel: func(i any) string {
				return fmt.Sprintf("level=%q", i)
			},
			FormatMessage: func(i any) string {
				return fmt.Sprintf("msg=%q", i) // Quoting the message automatically
			},
			NoColor: true,
		}
		output = zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	case "PRODUCTION":
		file, err := os.OpenFile("prod.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		output = zerolog.ConsoleWriter{
			Out:        file,
			TimeFormat: "",
			FormatFieldName: func(i any) string {
				return fmt.Sprintf("%s=", i)
			},
			FormatFieldValue: func(i any) string {
				s := fmt.Sprintf("%v", i)
				if strings.ContainsAny(s, " \t\n\r") {
					return fmt.Sprintf("%q", s)
				}
				return s
			},
			FormatTimestamp: func(i any) string {
				t, err := time.Parse(time.RFC3339, i.(string))
				if err != nil {
					return fmt.Sprintf("time=%q", i) // Fallback if parsing fails
				}
				return fmt.Sprintf("time=%d", t.UnixMilli())
			},
			FormatLevel: func(i any) string {
				return fmt.Sprintf("level=%s", i)
			},
			FormatMessage: func(i any) string {
				return fmt.Sprintf("msg=%q", i) // Quoting the message automatically
			},
			NoColor: true,
		}
	default:
		return nil, errors.New("invalid environment for logger setup")
	}

	logger := zerolog.New(output).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = time.RFC3339Nano
	return &LoggerService{
		Logger: logger,
		Env:    env,
	}, nil
}

func (l *LoggerService) Debug(msg string) {
	if l.Env == "PRODUCTION" {
		return
	}
	l.Logger.WithLevel(zerolog.DebugLevel).Msg(msg)
}

func (l *LoggerService) Info(msg string) {
	l.Logger.WithLevel(zerolog.InfoLevel).Msg(msg)
}

func (l *LoggerService) Warn(msg string) {
	l.Logger.WithLevel(zerolog.InfoLevel).Msg(msg)
}

func (l *LoggerService) Error(msg string, err error) {
	l.Logger.WithLevel(zerolog.InfoLevel).Err(err).Msg(msg)
}

func (l *LoggerService) Fatal(msg string, err error) {
	l.Logger.WithLevel(zerolog.FatalLevel).Err(err).Msg(msg)
}
