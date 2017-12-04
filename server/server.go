package server

import (
	"context"
	"io"
	"math"
	"time"

	"github.com/golang/protobuf/proto"

	"gitlab.com/ethanlewis787/fun-with-grpc/protos"
)

// RouteGuideServerImpl - implements the gRPC RouteGuideServer interface
type RouteGuideServerImpl struct {
	savedFeatures []*protos.Feature
	routeNotes    map[string][]*protos.RouteNote
}

// GetFeature returns the feature at the given point (simple RPC)
// The method is passed a context object for the RPC and the client's Point protocol buffer
// request. It returns a feature protocol buffer object with the response information and error
// In the method we populate the feature with the appropriate information and then return it along with
// an nil error to tell gRPC that we've finished dealing  with the RPC and that the feature can be returned
// to the client.
func (s *RouteGuideServerImpl) GetFeature(ctx context.Context, point *protos.Point) (*protos.Feature, error) {
	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}
	return &protos.Feature{Location: point}, nil
}

// ListFeatures lists all features contained within the given bounding rectangle (server side streaming)
// In the mehtod, we populate as many Feature objects as we need to return, writing them to the
// RouteGuide_ListFeaturesServer using its Send() method. Finally as in our simple RPC we return a nil error
// to rell gRPC that we've finsihed writing responses. Should any error happen in this call, we return a non-nil error
// The gRPC layer will transalte it into an appropriate RPC status to be sent on the wire.
func (s *RouteGuideServerImpl) ListFeatures(rect *protos.Rectangle, stream protos.RouteGuide_ListFeaturesServer) error {
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rect) {
			if err := stream.Send(feature); err != nil {
				return err
			}
		}
	}
	return nil
}

// RecordRoute records a route composited of a sequence of points. (client side streaming)
// It gets a stream of points, and responds with statistics about the "trip":
// number of points,  number of known features visited, total distance traveled, and
// total time spent.
// note : client side streaming is a little abstract for me. The server stream can Recv()
// build up the response then send it back to the client and close. It is abstract because this
// function assume that the reader knows the properties for stream. Their is no direct input/oput
// defined by this function. It is rather defined in the stream itself. It might be worth
// addign the rpc def in the comments for an example.
// i.e ( rpc RecordRoute(stream Point) returns (RouteSummary) {} ) <- less abstract :D
func (s *RouteGuideServerImpl) RecordRoute(stream protos.RouteGuide_RecordRouteServer) error {
	// Construct points for RouteSummary ( which is the return object )
	var pointCount, featureCount, distance int32
	var lastPoint *protos.Point
	startTime := time.Now()
	for {
		// get a point
		point, err := stream.Recv()
		// We are at the end of the stream
		if err == io.EOF {
			endTime := time.Now()
			// send summary and close the stream
			err := stream.SendAndClose(&protos.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
			// gRPC layer will handle status code if this is non-nil
			return err
		}
		if err != nil {
			return err
		}
		pointCount++
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}
		if lastPoint != nil {
			distance += calcDistance(lastPoint, point)
		}
		lastPoint = point
	}
}

// ------ Unexported helpers ------ //

// inRange checks if point is in bounds of Rectangle
func inRange(point *pb.Point, rect *pb.Rectangle) bool {
	left := math.Min(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	right := math.Max(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	top := math.Max(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
	bottom := math.Min(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))

	if float64(point.Longitude) >= left &&
		float64(point.Longitude) <= right &&
		float64(point.Latitude) >= bottom &&
		float64(point.Latitude) <= top {
		return true
	}
	return false
}

// toRadians converts a number to radian
func toRadians(num float64) float64 {
	return num * math.Pi / float64(180)
}

// calcDistance calculates the distance between two points using the "haversine" formula.
// This code was taken from http://www.movable-type.co.uk/scripts/latlong.html.
func calcDistance(p1 *pb.Point, p2 *pb.Point) int32 {
	const CordFactor float64 = 1e7
	const R float64 = float64(6371000) // metres
	lat1 := float64(p1.Latitude) / CordFactor
	lat2 := float64(p2.Latitude) / CordFactor
	lng1 := float64(p1.Longitude) / CordFactor
	lng2 := float64(p2.Longitude) / CordFactor
	φ1 := toRadians(lat1)
	φ2 := toRadians(lat2)
	Δφ := toRadians(lat2 - lat1)
	Δλ := toRadians(lng2 - lng1)

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return int32(distance)
}
