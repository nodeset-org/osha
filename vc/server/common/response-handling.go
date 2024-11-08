package common

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/validator/keymanager"
)

// Handle routes called with an invalid method
func HandleInvalidMethod(w http.ResponseWriter, logger *slog.Logger) {
	writeResponse(w, logger, http.StatusMethodNotAllowed, []byte{})
}

// Handles an error related to parsing the input parameters of a request
func HandleInputError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Write an error if the auth header couldn't be decoded
func HandleAuthHeaderError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg)
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the auth header is missing
func HandleMissingAuthHeader(w http.ResponseWriter, logger *slog.Logger) {
	msg := "No Authorization header found"
	bytes := formatError(msg)
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the auth header couldn't be decoded
func HandleServerError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg)
	writeResponse(w, logger, http.StatusInternalServerError, bytes)
}

// The request completed successfully
func HandleSuccess[DataType any](w http.ResponseWriter, logger *slog.Logger, data DataType, slashingProtection *beacon.SlashingProtectionData) {
	response := keymanager.KeyManagerResponse[DataType]{
		Data: data,
	}
	if slashingProtection != nil {
		response.SlashingProtection = *slashingProtection
	}

	// Serialize the response
	bytes, err := json.Marshal(response)
	if err != nil {
		HandleServerError(w, logger, fmt.Errorf("error serializing response: %w", err))
	}
	// Write it
	logger.Debug("Response body", slog.String(log.BodyKey, string(bytes)))
	writeResponse(w, logger, http.StatusOK, bytes)
}

// The request completed successfully and was accepted
func HandleAccepted(w http.ResponseWriter, logger *slog.Logger) {
	writeResponse(w, logger, http.StatusAccepted, nil)
}

// Writes a response to an HTTP request back to the client and logs it
func writeResponse(w http.ResponseWriter, logger *slog.Logger, statusCode int, message []byte) {
	// Prep the log attributes
	codeMsg := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	attrs := []any{
		slog.String(log.CodeKey, codeMsg),
	}

	// Log the response
	logMsg := "Responded with:"
	switch statusCode {
	case http.StatusOK, http.StatusAccepted:
		logger.Info(logMsg, attrs...)
	case http.StatusInternalServerError:
		logger.Error(logMsg, attrs...)
	default:
		logger.Warn(logMsg, attrs...)
	}

	// Write it to the client
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if message != nil {
		_, writeErr := w.Write(message)
		if writeErr != nil {
			logger.Error("Error writing response", "error", writeErr)
		}
	}
}

// JSONifies an error for responding to requests
func formatError(message string) []byte {
	msg := keymanager.KeyManagerResponse[struct{}]{
		Message: message,
	}

	bytes, _ := json.Marshal(msg)
	return bytes
}
