package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/goccy/go-json"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/osha/beacon/api"
	"github.com/nodeset-org/osha/beacon/db"
	"github.com/nodeset-org/osha/beacon/manager"
	"github.com/rocket-pool/node-manager-core/log"
)

type BeaconMockServer struct {
	logger  *slog.Logger
	ip      string
	port    uint16
	socket  net.Listener
	server  http.Server
	router  *mux.Router
	manager *manager.BeaconMockManager
}

func NewBeaconMockServer(logger *slog.Logger, ip string, port uint16, config *db.Config) (*BeaconMockServer, error) {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	server := &BeaconMockServer{
		logger: logger,
		ip:     ip,
		port:   port,
		router: router,
		server: http.Server{
			Handler: router,
		},
		manager: manager.NewBeaconMockManager(logger, config),
	}

	// Register each route
	apiRouter := router.PathPrefix("/eth").Subrouter()
	server.registerApiRoutes(apiRouter)
	adminRouter := router.PathPrefix("/admin").Subrouter()
	server.registerAdminRoutes(adminRouter)
	return server, nil
}

// Starts listening for incoming HTTP requests
func (s *BeaconMockServer) Start(wg *sync.WaitGroup) error {
	// Create the socket
	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	s.socket = socket

	// Get the port if random
	if s.port == 0 {
		s.port = uint16(socket.Addr().(*net.TCPAddr).Port)
	}

	// Start listening
	wg.Add(1)
	go func() {
		err := s.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error while listening for HTTP requests", log.Err(err))
		}
		wg.Done()
	}()

	return nil
}

// Stops the HTTP listener
func (s *BeaconMockServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}

// Get the port the server is listening on
func (s *BeaconMockServer) GetPort() uint16 {
	return s.port
}

// API routes
func (s *BeaconMockServer) registerApiRoutes(apiRouter *mux.Router) {
	apiRouter.HandleFunc("/"+api.ValidatorsRoute, s.getValidators)
	apiRouter.HandleFunc("/"+api.ValidatorRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getValidator(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	apiRouter.HandleFunc("/"+api.SyncingRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getSyncStatus(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	apiRouter.HandleFunc("/"+api.DepositContractRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getDepositContract(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	apiRouter.HandleFunc("/"+api.BeaconGenesisRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getBeaconGenesis(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	apiRouter.HandleFunc("/"+api.ConfigSpecRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getConfigSpec(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	apiRouter.HandleFunc("/"+api.FinalityCheckpointsRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getFinalityCheckpoints(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
}

// Admin routes
func (s *BeaconMockServer) registerAdminRoutes(adminRouter *mux.Router) {
	adminRouter.HandleFunc("/"+api.AddValidatorRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.addValidator(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	adminRouter.HandleFunc("/"+api.CommitBlockRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.commitBlock(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	adminRouter.HandleFunc("/"+api.SetBalanceRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.setBalance(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	adminRouter.HandleFunc("/"+api.SetStatusRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.setStatus(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	adminRouter.HandleFunc("/"+api.SlashRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.slash(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
	adminRouter.HandleFunc("/"+api.SetHighestSlotRoute, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.setHighestSlot(w, r)
		default:
			handleInvalidMethod(s.logger, w)
		}
	})
}

// =============
// === Utils ===
// =============

func (s *BeaconMockServer) processApiRequest(w http.ResponseWriter, r *http.Request, requestBody any) url.Values {
	args := r.URL.Query()
	s.logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
	s.logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

	if requestBody != nil {
		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			handleInputError(s.logger, w, fmt.Errorf("error reading request body: %w", err))
			return nil
		}
		s.logger.Debug("Request body:", slog.String(log.BodyKey, string(bodyBytes)))

		// Deserialize the body
		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			handleInputError(s.logger, w, fmt.Errorf("error deserializing request body: %w", err))
			return nil
		}
	}

	return args
}
