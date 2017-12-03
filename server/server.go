package server

import (
	"context"
	"math"

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

// RecordRoute -
func (s *RouteGuideServerImpl) RecordRoute(stream protos.RouteGuide_RecordRouteServer) error {

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
