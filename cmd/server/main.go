package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/contextkeys"
	"github.com/tilseiffert/docker-tex-to-pdf/internal/restserver"
	server "github.com/tilseiffert/docker-tex-to-pdf/pkg/server"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"
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

func RegisterEndpoints(muxer *http.ServeMux) {

	path := "/api/v1/"

	muxer.HandleFunc(path+"ping", func(w http.ResponseWriter, r *http.Request) {

		logger := r.Context().Value("logger").(*slog.Logger)
		logger.Debug("Hello from /ping handler")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	muxer.HandleFunc(path+"health", func(w http.ResponseWriter, r *http.Request) {

		logger := r.Context().Value("logger").(*slog.Logger)
		logger.Debug("Hello from /health handler")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func main() {

	zerologLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	logger := slog.New(slogzerolog.Option{Level: slog.LevelDebug, Logger: &zerologLogger}.NewZerologHandler())
	logger.Debug("Hello World 👋")

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	if err != nil {
		logger.Error("failed to connect database", "error", err)
	}

	apiserver, err := restserver.NewServer(db, &restserver.ServerOptions{
		BUILDDIR_TEMPLATE: "tex-to-pdfa.*.build",
	})

	if err != nil {
		logger.Error("failed to create api-server", "error", err)
	}

	srv := server.NewRestServer(logger)

	opts := server.RestServerOptions{
		Address:                  "localhost:6204",
		OptLogReqeust:            true,
		CallbackEndpointRegister: apiserver.RegisterEndpoints,
	}

	err = srv.Start(&opts)

	if err != nil {
		logger.Error("failed to start server", "error", err)
	}
}
