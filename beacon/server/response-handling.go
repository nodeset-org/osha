package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/osha/beacon/api"
	"github.com/rocket-pool/node-manager-core/log"
)

// Handle routes called with an invalid method
func handleInvalidMethod(logger *slog.Logger, w http.ResponseWriter) {
	writeResponse(logger, w, http.StatusMethodNotAllowed, []byte{})
}

// Handles an error related to parsing the input parameters of a request
func handleInputError(logger *slog.Logger, w http.ResponseWriter, err error) {
	msg := err.Error()
	code := http.StatusBadRequest
	bytes := formatError(code, msg)
	writeResponse(logger, w, code, bytes)
}

// Write an error if the auth header couldn't be decoded
func handleServerError(logger *slog.Logger, w http.ResponseWriter, err error) {
	msg := err.Error()
	code := http.StatusInternalServerError
	bytes := formatError(code, msg)
	writeResponse(logger, w, code, bytes)
}

// The request completed successfully
func handleSuccess(logger *slog.Logger, w http.ResponseWriter, message any) {
	bytes := []byte{}
	if message != nil {
		// Serialize the response
		var err error
		bytes, err = json.Marshal(message)
		if err != nil {
			handleServerError(logger, w, fmt.Errorf("error serializing response: %w", err))
		}
	}

	// Write it
	logger.Debug("Response body", slog.String(log.BodyKey, string(bytes)))
	writeResponse(logger, w, http.StatusOK, bytes)
}

// Writes a response to an HTTP request back to the client and logs it
func writeResponse(logger *slog.Logger, w http.ResponseWriter, statusCode int, message []byte) {
	// Prep the log attributes
	codeMsg := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	attrs := []any{
		slog.String(log.CodeKey, codeMsg),
	}

	// Log the response
	logMsg := "Responded with:"
	switch statusCode {
	case http.StatusOK:
		logger.Info(logMsg, attrs...)
	case http.StatusInternalServerError:
		logger.Error(logMsg, attrs...)
	default:
		logger.Warn(logMsg, attrs...)
	}

	// Write it to the client
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(message)
	if writeErr != nil {
		logger.Error("Error writing response", "error", writeErr)
	}
}

// JSONifies an error for responding to requests
func formatError(code int, message string) []byte {
	msg := api.ErrorResponse{
		Code:    code,
		Message: message,
	}

	bytes, _ := json.Marshal(msg)
	return bytes
}
