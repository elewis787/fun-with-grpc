## gRPC example golang 
This is pretty much a copy of https://github.com/grpc/grpc-go/tree/master/examples/route_guide and https://grpc.io/docs/tutorials/basic/go.html#why-use-grpc

I changed the layout for readability

## Fun with gRPC
Playing around with googles grpc for fun. 


### Generate the service and protobuffers

`protoc -I protos/ protos/route_guide.proto --go_out=plugins=grpc:protos` 



