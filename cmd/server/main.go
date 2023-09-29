package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/contextkeys"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/server"
)

// var loggerKey = &contextKey{"logger"}

// initLogger initializes a logger and adds it to the context
func initLogger(ctx context.Context) context.Context {

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Caller().Logger()

	return context.WithValue(ctx, contextkeys.LoggerKey, logger)
}

// Log returns the logger from the context (for more convenient logging)
func Log(ctx context.Context) *zerolog.Logger {

	if logger, ok := ctx.Value(contextkeys.LoggerKey).(zerolog.Logger); ok {
		return &logger
	}

	panic("unable to retrieve logger from context")
}

func main() {

	// === Initialize context ===

	starttime := time.Now()

	// Initialize context
	ctx := initLogger(context.Background())

	Log(ctx).Debug().Msg("Hello world 👋")

	s, err := server.Start(server.StandardPort)

	if err != nil {
		Log(ctx).Fatal().Err(err).Msg("Error starting server")
	}

	Log(ctx).Info().Str("serverinfo", fmt.Sprint(s.GetServiceInfo())).Msg("Server started")

	// === Tidy up ===

	// log runtime
	Log(ctx).Debug().Str("runtime", time.Since(starttime).String()).Msg("Done 👋")
}
