package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"gitlab.com/ethanlewis787/fun-with-grpc/protos"
	"gitlab.com/ethanlewis787/fun-with-grpc/server"

	"go.uber.org/zap"

	"github.com/urfave/cli"
)

func main() {
	appConfig := &config{}
	// Init the urfave cli app
	app := cli.NewApp()
	app.Name = "epidemic-cli"
	app.Usage = "cli used to interact with a gRPC server"
	app.Version = "v0.0.0" // major,minor,patch
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "port",
			Value:       "10101", // default value
			Usage:       "gRPC server port",
			EnvVar:      "port",
			Destination: &appConfig.gRCPPort,
		},
		cli.StringFlag{
			Name:        "file-path",
			Value:       "./testdata/route_guide_db.json", // default value
			Usage:       "test data file",
			EnvVar:      "file-path",
			Destination: &appConfig.filePath,
		},
	} // defined in flags.go
	// ------- Main Application function -------
	app.Action = func(cliCTX *cli.Context) error {
		// Init zap logger
		zlogger, err := zap.NewDevelopment() // TODO this needs to be configured for production
		if err != nil {
			return err
		}
		zlogger.Info("creating grpc server")

		rs := new(server.RouteGuideServerImpl)
		rs.LoadFeatures(appConfig.filePath)
		rs.RouteNotes = make(map[string][]*protos.RouteNote)

		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
		protos.RegisterRouteGuideServer(grpcServer, rs)
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", appConfig.gRCPPort))
		if err != nil {
			zlogger.Error("failed to listen")
			return err
		}
		zlogger.Info("serving")
		grpcServer.Serve(lis)

		return nil
	}
	// Start main
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
