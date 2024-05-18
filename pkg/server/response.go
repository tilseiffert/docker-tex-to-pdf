package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type CommonResponse struct {
	Status  int         `json:"status"`  // HTTP status code
	Message string      `json:"message"` // Human-readable message
	Data    interface{} `json:"data"`    // Data payload
}

func writeActualResponse(w http.ResponseWriter, resp CommonResponse, logger *slog.Logger) error {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Status)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errMsg := "failed to encode/write response [LLZHGR08]"
		http.Error(w, errMsg, http.StatusInternalServerError)

		if logger != nil {
			logger.Error(errMsg, "error", err)
		}

		return fmt.Errorf(errMsg+": %w", err)
	}

	return nil
}

// WriteResponse writes a response to the client with the given data. The response will have a status code of 200 (OK).
// With optional logger to log the response message, may be nil. If not nil, the logger will be used to log the response message at debug level. All returned errors will also be logged.
func WriteResponse(w http.ResponseWriter, data interface{}, logger *slog.Logger) error {

	code := http.StatusOK

	resp := CommonResponse{
		Status:  code,
		Message: http.StatusText(code),
		Data:    data,
	}

	if logger != nil {
		logger.Debug("sending response "+fmt.Sprint(code), "data", data)
	}

	return writeActualResponse(w, resp, logger)
}

// Error replies to the request with the specified error message and HTTP code. It does not otherwise end the request; the caller should ensure no further writes are done to w. The error message should be plain text.
// Param msg is an optional message to be sent to the client, it will be prefixed with the http error message.
// Param logger is an optional logger to log the error message, may be nil. If not nil, the logger will be used to log the error message at ubfi level. All returned errors will also be logged.
func WriteError(w http.ResponseWriter, code int, msg string, logger *slog.Logger) error {

	if msg == "" {
		msg = http.StatusText(code)
	} else {
		msg = http.StatusText(code) + ": " + msg
	}

	resp := CommonResponse{
		Status:  code,
		Message: msg,
		Data:    nil,
	}

	if logger != nil {
		logger.Info("sending error response "+fmt.Sprint(code), "error-message", msg)
	}

	return writeActualResponse(w, resp, logger)
}
