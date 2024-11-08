package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	beacondb "github.com/nodeset-org/osha/beacon/db"
	"github.com/nodeset-org/osha/vc/db"
	"github.com/nodeset-org/osha/vc/server"
	"github.com/urfave/cli/v2"
)

const (
	Version string = "0.1.0"
)

// Run
func main() {
	// Initialise application
	app := cli.NewApp()

	// Set application info
	app.Name = "osha-bn"
	app.Usage = "Partial mock of a Validator Client, useful for testing applications that use the key manager API"
	app.Version = Version
	app.Authors = []*cli.Author{
		{
			Name:  "NodeSet",
			Email: "info@nodeset.io",
		},
	}
	app.Copyright = "(C) 2024 NodeSet LLC"

	ipFlag := &cli.StringFlag{
		Name:    "ip",
		Aliases: []string{"i"},
		Usage:   "The IP address to bind the API server to",
		Value:   "127.0.0.1",
	}
	portFlag := &cli.UintFlag{
		Name:    "port",
		Aliases: []string{"p"},
		Usage:   "The port to bind the API server to",
		Value:   48822,
	}
	defaultFeeRecipientFlag := &cli.StringFlag{
		Name:    "default-fee-recipient",
		Aliases: []string{"f"},
		Usage:   "The default fee recipient address",
		Value:   db.DefaultFeeRecipientString,
	}
	defaultGraffitiFlag := &cli.StringFlag{
		Name:    "default-graffiti",
		Aliases: []string{"g"},
		Usage:   "The default graffiti string",
		Value:   db.DefaultGraffiti,
	}
	genesisValidatorsRootFlag := &cli.StringFlag{
		Name:    "genesis-validators-root",
		Aliases: []string{"r"},
		Usage:   "The genesis validators root hash",
		Value:   beacondb.DefaultGenesisValidatorsRootString,
	}
	jwtSecretFlag := &cli.StringFlag{
		Name:    "jwt-secret",
		Aliases: []string{"j"},
		Usage:   "The JWT secret",
		Value:   db.DefaultJwtSecret,
	}

	app.Flags = []cli.Flag{
		ipFlag,
		portFlag,
		defaultFeeRecipientFlag,
		defaultGraffitiFlag,
		genesisValidatorsRootFlag,
		jwtSecretFlag,
	}
	app.Action = func(c *cli.Context) error {
		logger := slog.Default()

		// Create the server
		var err error
		ip := c.String(ipFlag.Name)
		port := uint16(c.Uint(portFlag.Name))
		defaultFeeRecipient := common.HexToAddress(c.String(defaultFeeRecipientFlag.Name))
		defaultGraffiti := c.String(defaultGraffitiFlag.Name)
		genesisValidatorsRoot := common.HexToHash(c.String(genesisValidatorsRootFlag.Name))
		jwtSecret := c.String(jwtSecretFlag.Name)
		server, err := server.NewVcMockServer(logger, ip, port, db.KeyManagerDatabaseOptions{
			DefaultFeeRecipient:   &defaultFeeRecipient,
			DefaultGraffiti:       &defaultGraffiti,
			GenesisValidatorsRoot: &genesisValidatorsRoot,
			JwtSecret:             &jwtSecret,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating server: %v", err)
			os.Exit(1)
		}

		// Start it
		wg := &sync.WaitGroup{}
		err = server.Start(wg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting server: %v", err)
			os.Exit(1)
		}
		port = server.GetPort()

		// Handle process closures
		termListener := make(chan os.Signal, 1)
		signal.Notify(termListener, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-termListener
			fmt.Println("Shutting down...")
			err := server.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error stopping server: %v", err)
				os.Exit(1)
			}
		}()

		// Run the daemon until closed
		logger.Info(fmt.Sprintf("Started OSHA VC mock server on %s:%d", ip, port))
		wg.Wait()
		fmt.Println("Server stopped.")
		return nil
	}

	// Run application
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
