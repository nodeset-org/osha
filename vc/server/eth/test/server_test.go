package eth_test

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/nodeset-org/osha/vc/db"
	"github.com/nodeset-org/osha/vc/manager"
	"github.com/nodeset-org/osha/vc/server"
)

// Various singleton variables used for testing
var (
	logger      *slog.Logger           = slog.Default()
	s           *server.VcMockServer   = nil
	mgr         *manager.VcMockManager = nil
	wg          *sync.WaitGroup        = nil
	port        uint16                 = 0
	jwtFilePath string
)

// Initialize a common server used by all tests
func TestMain(m *testing.M) {
	// Create the server
	var err error
	s, err = server.NewVcMockServer(logger, "localhost", 0, db.KeyManagerDatabaseOptions{})
	if err != nil {
		fail("error creating server: %v", err)
	}
	logger.Info("Created server")

	// Write the JWT file
	jwtFile, err := os.CreateTemp("", "osha-jwt-*.txt")
	if err != nil {
		fail("error creating JWT file: %v", err)
	}
	_, err = jwtFile.WriteString(db.DefaultJwtSecret)
	if err != nil {
		fail("error writing JWT file: %v", err)
	}
	jwtFilePath = jwtFile.Name()
	logger.Info("Created JWT file", "path", jwtFilePath)

	// Start it
	wg = &sync.WaitGroup{}
	err = s.Start(wg)
	if err != nil {
		fail("error starting server: %v", err)
	}
	port = s.GetPort()
	logger.Info("Started server", "port", port)
	mgr = s.GetManager()

	// Run tests
	code := m.Run()

	// Revert to the baseline after testing is done
	cleanup()

	// Done
	os.Exit(code)
}

func fail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logger.Error(msg)
	cleanup()
	os.Exit(1)
}

func cleanup() {
	if s != nil {
		_ = s.Stop()
		wg.Wait()
		logger.Info("Stopped server")
	}
	if jwtFilePath != "" {
		_ = os.Remove(jwtFilePath)
		logger.Info("Removed JWT file")
	}
}
