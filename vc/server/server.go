package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/osha/vc/db"
	"github.com/nodeset-org/osha/vc/manager"
	"github.com/nodeset-org/osha/vc/server/admin"
	"github.com/nodeset-org/osha/vc/server/eth"
	"github.com/rocket-pool/node-manager-core/log"
)

type VcMockServer struct {
	logger  *slog.Logger
	ip      string
	port    uint16
	socket  net.Listener
	server  http.Server
	router  *mux.Router
	manager *manager.VcMockManager

	// Route handlers
	adminServer *admin.AdminServer
	ethServer   *eth.KeyManagerServer
}

func NewVcMockServer(logger *slog.Logger, ip string, port uint16, dbOpts db.KeyManagerDatabaseOptions) (*VcMockServer, error) {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	server := &VcMockServer{
		logger: logger,
		ip:     ip,
		port:   port,
		router: router,
		server: http.Server{
			Handler: router,
		},
		manager: manager.NewVcMockManager(logger, dbOpts),
	}
	server.adminServer = admin.NewAdminServer(logger, server.manager)
	server.ethServer = eth.NewKeyManagerServer(logger, server.manager)

	// Register admin routes
	adminRouter := router.PathPrefix("/admin").Subrouter()
	server.adminServer.RegisterRoutes(adminRouter)

	// Register API routes
	apiRouter := router.PathPrefix("/eth").Subrouter()
	server.ethServer.RegisterRoutes(apiRouter)

	return server, nil
}

// Starts listening for incoming HTTP requests
func (s *VcMockServer) Start(wg *sync.WaitGroup) error {
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
func (s *VcMockServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}

// Get the port the server is listening on
func (s *VcMockServer) GetPort() uint16 {
	return s.port
}

// Get the mock manager for direct access
func (s *VcMockServer) GetManager() *manager.VcMockManager {
	return s.manager
}
