package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Default = zerolog.New(os.Stderr).
	Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).
	Level(zerolog.InfoLevel).
	With().
	Timestamp().
	Caller().
	Logger()
