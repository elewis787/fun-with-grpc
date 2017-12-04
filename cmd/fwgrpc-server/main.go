package main

import (
	"log"
	"os"

	"go.uber.org/zap"

	"github.com/urfave/cli"
)

func main() {
	// Init the urfave cli app
	app := cli.NewApp()
	app.Name = "epidemic-cli"
	app.Usage = "cli used to interact with a gRPC server"
	app.Version = "v0.0.0"   // major,minor,patch
	app.Flags = []cli.Flag{} // defined in flags.go
	// ------- Main Application function -------
	app.Action = func(cliCTX *cli.Context) error {
		// Init zap logger
		zlogger, err := zap.NewDevelopment() // TODO this needs to be configured for production
		if err != nil {
			return err
		}
		zlogger.Info("epic")
		return nil
	}
	// Start main
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
