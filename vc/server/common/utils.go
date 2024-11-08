package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/log"
)

// ==============
// === Errors ===
// ==============

var (
	ErrInvalidSession error = errors.New("session token is invalid")
)

// Logs the request and returns the query args and path args
func ProcessApiRequest(serverImpl IServerImpl, w http.ResponseWriter, r *http.Request, requestBody any) (url.Values, map[string]string, bool) {
	args := r.URL.Query()
	logger := serverImpl.GetLogger()
	logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
	logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

	if requestBody != nil {
		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			HandleInputError(w, logger, fmt.Errorf("error reading request body: %w", err))
			return nil, nil, false
		}
		logger.Debug("Request body:", slog.String(log.BodyKey, string(bodyBytes)))

		// Deserialize the body
		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			HandleInputError(w, logger, fmt.Errorf("error deserializing request body: %w", err))
			return nil, nil, false
		}
	}

	return args, mux.Vars(r), true
}

// Makes sure the request has a valid auth header and returns the session it belongs to
func ProcessAuthHeader(serverImpl IServerImpl, w http.ResponseWriter, r *http.Request) bool {
	// Get the auth header
	mgr := serverImpl.GetManager()
	logger := serverImpl.GetLogger()

	// Get the bearer
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		HandleMissingAuthHeader(w, logger)
		return false
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		HandleAuthHeaderError(w, logger, fmt.Errorf("invalid auth header: %s", authHeader))
		return false
	}

	// Make sure the token is correct
	token := strings.TrimPrefix(authHeader, "Bearer ")
	db := mgr.GetDatabase()
	expectedToken := db.GetJwtSecret()
	if token != expectedToken {
		HandleAuthHeaderError(w, logger, fmt.Errorf("invalid auth token: %s", token))
		return false
	}
	return true
}
