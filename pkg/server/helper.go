package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (srv *RestServer) logRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		requestID, err := srv.NewULID()
		w.Header().Set("X-Request-ID", requestID.String())

		if err != nil {
			srv.Logger.Error("failed to generate request ID", "error-id", "1AMDKTTM", "error", err)
		}

		logger := srv.Logger.With("requestID", requestID)

		logger.Info(fmt.Sprintf("REQUEST: %-4s %s", r.Method, r.URL.Path),
			"method", r.Method,
			"url", r.URL.Path,
			"host", r.Host,
		)

		ctx := context.WithValue(r.Context(), "logger", logger)
		ctx = context.WithValue(ctx, "requestID", requestID)
		r = r.WithContext(ctx)

		// ctxLogger, ok := r.Context().Value("logger").(*slog.Logger)
		// if !ok {
		// 	h.logger.Error("failed to extract logger from context", "error-id", "336TICKQ")
		// 	return
		// }

		// Rufe den nächsten Handler in der Kette auf
		next.ServeHTTP(w, r)

		stopTime := time.Now()

		logger.Debug(fmt.Sprintf("request completed in %s", stopTime.Sub(startTime).String()),
			"duration", stopTime.Sub(startTime),
		)

		// h.logger.Debug(fmt.Sprintf("RESPONSE: %-4s %s", r.Method, r.URL.Path))
	})
}
