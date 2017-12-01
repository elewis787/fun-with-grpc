## Fun with gRPC
Playing around with googles grpc for fun. 


### Generate the service and protobuffers

`protoc -I protos/ protos/route_guide.proto --go_out=plugins=grpc:protos` 



