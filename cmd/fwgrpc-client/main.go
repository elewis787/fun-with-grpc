package main

import (
	"log"
	"os"

	"golang.org/x/net/context"

	"gitlab.com/ethanlewis787/fun-with-grpc/client"
	"gitlab.com/ethanlewis787/fun-with-grpc/protos"

	"google.golang.org/grpc"

	"go.uber.org/zap"

	"github.com/urfave/cli"
)

func main() {
	appConfig := &config{}
	// Init the urfave cli app
	app := cli.NewApp()
	app.Name = "fun-with-grpc-client"
	app.Usage = "cli used to interact with a gRPC client"
	app.Version = "v0.0.0" // major,minor,patch
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "server-address",
			Value:       "127.0.0.1:10101", // default value
			Usage:       "gRPC server address and port",
			EnvVar:      "SERVER_ADDRESS",
			Destination: &appConfig.gRPCServerAddr,
		},
	} // defined in flags.go
	// ------- Main Application function -------
	app.Action = func(cliCTX *cli.Context) error {
		// Init zap logger
		zlogger, err := zap.NewDevelopment() // TODO this needs to be configured for production
		if err != nil {
			return err
		}

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())

		conn, err := grpc.Dial(appConfig.gRPCServerAddr, opts...)
		if err != nil {
			zlogger.Error("fail to dial :", zap.Error(err))
			return err
		}
		defer conn.Close()

		grpcClient := protos.NewRouteGuideClient(conn)

		routeClient := &client.Client{
			RouteGuideClient: grpcClient,
			Zlogger:          zlogger,
		}

		// looking for valid
		err = routeClient.PrintFeature(context.Background(), &protos.Point{Latitude: 409146138, Longitude: -746188906})
		if err != nil {
			zlogger.Error("got", zap.Error(err))
		}
		// looking for missing
		err = routeClient.PrintFeature(context.Background(), &protos.Point{Latitude: 0, Longitude: 0})
		if err != nil {
			zlogger.Error("got", zap.Error(err))
		}

		err = routeClient.PrintFeatures(context.Background(), &protos.Rectangle{
			Lo: &protos.Point{Latitude: 400000000, Longitude: -750000000},
			Hi: &protos.Point{Latitude: 420000000, Longitude: -730000000},
		})
		if err != nil {
			zlogger.Error("got", zap.Error(err))
		}

		err = routeClient.RunRecordRoute(context.Background())
		if err != nil {
			zlogger.Error("got", zap.Error(err))
		}

		err = routeClient.RunRouteChat(context.Background())
		if err != nil {
			zlogger.Error("got", zap.Error(err))
		}

		return nil
	}
	// Start main
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
