package client

import (
	"io"
	"math/rand"
	"sync"
	"time"

	"gitlab.com/ethanlewis787/fun-with-grpc/protos"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Client wrapper for RouteGuideClient
type Client struct {
	RouteGuideClient protos.RouteGuideClient
	Zlogger          *zap.Logger
}

// PrintFeature - get the feature for the given point
func (c *Client) PrintFeature(ctx context.Context, point *protos.Point) error {
	feature, err := c.RouteGuideClient.GetFeature(ctx, point)
	if err != nil {
		return err
	}
	c.Zlogger.Info("Found", zap.Any("feature", feature))
	return nil
}

// PrintFeatures - get a list of features within the given bouding rectangle
func (c *Client) PrintFeatures(ctx context.Context, rect *protos.Rectangle) error {
	c.Zlogger.Info("Looking for features within : ", zap.Any("rect", rect))
	stream, err := c.RouteGuideClient.ListFeatures(ctx, rect)
	if err != nil {
		return err
	}
	// not a big fan of this forever loop stuff wish Recv would block or return eof :/
	for {
		feature, err := stream.Recv()
		// at the end
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		c.Zlogger.Info("Found", zap.Any("feature", feature))
	}
	return nil
}

// RunRecordRoute sends a sequence of points to server and expects to get a RouteSummary from server.
func (c *Client) RunRecordRoute(ctx context.Context) error {
	// create a random number of random points
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	pointCount := int(r.Int31n(100)) + 2 // traverse at least two points
	var points []*protos.Point
	for i := 0; i < pointCount; i++ {
		points = append(points, randomPoint(r))
	}
	c.Zlogger.Info("traversing points : ", zap.Int("length", len(points)))
	stream, err := c.RouteGuideClient.RecordRoute(ctx)
	if err != nil {
		return err
	}
	for _, point := range points {
		if err := stream.Send(point); err != nil {
			return err
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	c.Zlogger.Info("route summary", zap.Any("summary", reply))
	return nil
}

// RunRouteChat - receives a sequence of route notes, while sending notes for various locations.
func (c *Client) RunRouteChat(ctx context.Context) error {

	notes := []*protos.RouteNote{
		{&protos.Point{Latitude: 0, Longitude: 1}, "First message"},
		{&protos.Point{Latitude: 0, Longitude: 2}, "Second message"},
		{&protos.Point{Latitude: 0, Longitude: 3}, "Third message"},
		{&protos.Point{Latitude: 0, Longitude: 1}, "Fourth message"},
		{&protos.Point{Latitude: 0, Longitude: 2}, "Fifth message"},
		{&protos.Point{Latitude: 0, Longitude: 3}, "Sixth message"},
	}
	stream, err := c.RouteGuideClient.RouteChat(ctx)
	if err != nil {
		return err
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				wg.Done()
				return
			}
			if err != nil {
				c.Zlogger.Error("failed to receive note :", zap.Error(err))
			}
			c.Zlogger.Info("Got message", zap.String("message", in.Message),
				zap.Int32("lat", in.Location.Latitude), zap.Int32("long", in.Location.Longitude))
		}
	}()
	for _, note := range notes {
		c.Zlogger.Info("sending", zap.Any("note", note))
		if err := stream.Send(note); err != nil {
			return err
		}
	}
	stream.CloseSend()
	wg.Wait()
	return nil
}

// ------ Unexported helpers ------ //

// randomPoint return an random point that meets the lat/long requirements.
func randomPoint(r *rand.Rand) *protos.Point {
	lat := (r.Int31n(180) - 90) * 1e7
	long := (r.Int31n(360) - 180) * 1e7
	return &protos.Point{Latitude: lat, Longitude: long}
}
