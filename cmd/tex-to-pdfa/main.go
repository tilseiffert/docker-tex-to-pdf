package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/contextkeys"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/textopdfa"
)

const (
	BUILDDIR_TEMPLATE = "tex-to-pdfa_build_*"
	TEXFILE           = "main.tex"
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

	// === Compile TeX to PDF/A ===
	result, err := textopdfa.CompileTexToPDFA(ctx, TEXFILE, BUILDDIR_TEMPLATE)

	if err != nil {
		Log(ctx).Fatal().Err(err).Msg("Error compiling TeX to PDF/A")
	}

	Log(ctx).Info().Str("path", result).Msg("Successfully compiled TeX to PDF/A")

	// === Tidy up ===

	// log runtime
	Log(ctx).Debug().Str("runtime", time.Since(starttime).String()).Msg("Done 👋")
}
